package client

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	context "golang.org/x/net/context"

	"sync"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	"github.com/olekukonko/tablewriter"
	"github.com/xeonx/timeago"
)

// MaxBulkObjCount determines the number of bulk objects in commands that perform bulk requests
var MaxBulkObjCount = 25

// CreateCocoon a new cocoon
func CreateCocoon(cocoonPayload *proto_api.CocoonReleasePayloadRequest) error {

	ss := util.Spinner("Please wait...")
	defer ss()

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
	defer cc()
	cocoonJSON, err := proto_api.NewAPIClient(conn).CreateCocoon(ctx, cocoonPayload)
	if err != nil {
		return err
	}

	var cocoon types.Cocoon
	util.FromJSON(cocoonJSON.Body, &cocoon)

	ss()

	log.Info(`==> New cocoon created`)
	log.Infof(`==> Cocoon ID:  %s`, cocoon.ID)
	log.Infof(`==> Release ID: %s`, cocoon.Releases[0])

	return nil
}

// UpdateCocoon updates a cocoon and optionally creates a new
// release. A new release is created when Release fields are
// set/defined. No release is created if updated release fields match
// existing fields.
func UpdateCocoon(id string, upd *proto_api.CocoonReleasePayloadRequest) error {

	stopSpinner := util.Spinner("Please wait")

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
	defer cc()
	cl := proto_api.NewAPIClient(conn)
	resp, err := cl.UpdateCocoon(ctx, upd)
	if err != nil {
		if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
			stopSpinner()
			return fmt.Errorf("Cocoon not found: %s", id)
		}
		stopSpinner()
		return err
	}

	stopSpinner()

	var body map[string]interface{}
	if err = util.FromJSON(resp.Body, &body); err != nil {
		return fmt.Errorf("failed to coerce to map")
	}

	// no new updates was detected or performed.
	if !body["cocoonUpdated"].(bool) && len(body["newReleaseID"].(string)) == 0 {
		log.Info("No change detected. Nothing to do")
		return nil
	}

	log.Info(`==> Update successfully applied`)

	if body["cocoonUpdated"].(bool) {
		log.Infof(`==> Cocoon has been updated`)
	}

	if len(body["newReleaseID"].(string)) != 0 {
		log.Infof(`==> New Release ID: %s`, body["newReleaseID"].(string))
	}

	return nil
}

// GetCocoons fetches one or more cocoons and logs them
func GetCocoons(ids []string) error {

	var cocoons = []types.Cocoon{}
	var err error
	var resp *proto_api.Response

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	for _, id := range ids {
		stopSpinner := util.Spinner("Please wait")

		ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
		defer cc()
		cl := proto_api.NewAPIClient(conn)
		resp, err = cl.GetCocoon(ctx, &proto_api.GetCocoonRequest{ID: id})
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				stopSpinner()
				err = fmt.Errorf("Cocoon not found: %s", id)
				break
			}
			stopSpinner()
			break
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
		stopSpinner()
	}

	bs, _ := json.MarshalIndent(cocoons, "", "   ")
	log.Infof("%s", bs)
	if err != nil {
		return err
	}

	return nil
}

// Deploy creates and sends a deploy request to the server
func deploy(ctx context.Context, cocoonID string, useLastDeployedReleaseID bool) error {

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	client := proto_api.NewAPIClient(conn)
	resp, err := client.Deploy(ctx, &proto_api.DeployRequest{
		CocoonID: cocoonID,
	})
	if err != nil {
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// ListCocoons fetches and displays running cocoons belonging to
// the logged in user. Set showAll to true to list both running
// and stopped cocoons.
func ListCocoons(showAll, jsonFormatted bool) error {

	var cocoons []types.Cocoon

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
	defer cc()
	client := proto_api.NewAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	userSession, err := GetUserSessionToken()
	if err != nil {
		return ErrNoUserSession
	}

	resp, err := client.GetIdentity(ctx, &proto_api.GetIdentityRequest{
		Email: userSession.Email,
	})
	if err != nil {
		stopSpinner()
		return err
	}

	var identity types.Identity
	if err = util.FromJSON(resp.Body, &identity); err != nil {
		stopSpinner()
		return common.JSONCoerceErr("identity", err)
	}

	for _, cid := range identity.Cocoons {

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		resp, err = client.GetCocoon(ctx, &proto_api.GetCocoonRequest{
			ID: cid,
		})
		if err != nil {
			stopSpinner()
			return err
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
	}

	stopSpinner()

	if jsonFormatted {
		bs, _ := json.MarshalIndent(cocoons, "", "   ")
		log.Info("%s", bs)
		return nil
	}

	tableData := [][]string{}
	for _, cocoon := range cocoons {
		if !showAll && !util.InStringSlice([]string{api.CocoonStatusStarted, api.CocoonStatusBuilding, api.CocoonStatusRunning}, cocoon.Status) {
			continue
		}
		created, _ := time.Parse(time.RFC3339, cocoon.CreatedAt)
		tableData = append(tableData, []string{
			cocoon.ID,
			cocoon.Releases[len(cocoon.Releases)-1],
			common.CapitalizeString(timeago.English.Format(created)),
			common.CapitalizeString(cocoon.Status),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COCOON ID", "RELEASE ID (LATEST)", "CREATED", "STATUS"})
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.AppendBulk(tableData)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.Render()

	return nil
}

// StopCocoon one or more running cocoon codes
func StopCocoon(ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	var errs []error
	var stopped []string

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	cl := proto_api.NewAPIClient(conn)
	stopSpinner := util.Spinner("Please wait")

	for _, id := range ids {

		ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
		defer cc()
		_, err = cl.StopCocoon(ctx, &proto_api.StopCocoonRequest{ID: id})
		if err != nil {
			stopSpinner()
			return err
		}

		stopped = append(stopped, id)
	}

	stopSpinner()

	for _, id := range stopped {
		log.Info(id)
	}
	for _, err := range errs {
		log.Infof("Err: %s", err.Error())
	}

	return nil
}

// Start starts one or more new or stopped cocoon code.
// If useLastDeployedReleaseID is set to true, the scheduler will use the
// most recently approved and deployed release, otherwise it will
// try to deploy the latest release.
func Start(ids []string, useLastDeployedReleaseID bool) error {

	var errs []error
	var started []string
	var muErr = sync.Mutex{}
	var muStarted = sync.Mutex{}
	var wg = sync.WaitGroup{}

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	stopSpinner := util.Spinner("Please wait")

	for _, id := range ids {
		id := id
		wg.Add(1)
		go func() {
			ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
			defer cc()
			if err := deploy(ctx, id, useLastDeployedReleaseID); err != nil {
				muErr.Lock()
				errs = append(errs, fmt.Errorf("%s: %s", common.GetShortID(id), common.GetRPCErrDesc(err)))
				muErr.Unlock()
			} else {
				muStarted.Lock()
				started = append(started, id)
				muStarted.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	stopSpinner()

	if len(started) > 0 {
		log.Info("==> Successfully deployed the following:")
		for i, id := range started {
			log.Infof("==> %d. %s", i+1, id)
		}
	}

	for _, err := range errs {
		log.Infof(common.GetRPCErrDesc(err))
	}

	return nil
}

// AddSignatories adds one or more valid identities to a cocoon's signatory list.
// All valid identities are included and invalid ones will produce an error log..
func AddSignatories(cocoonID string, ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
	defer cc()
	cl := proto_api.NewAPIClient(conn)
	resp, err := cl.AddSignatories(ctx, &proto_api.AddSignatoriesRequest{
		CocoonID: cocoonID,
		IDs:      ids,
	})

	stopSpinner()

	if err != nil {
		return err
	}

	var r map[string][]string
	util.FromJSON(resp.Body, &r)

	if len(r["added"]) == 0 {
		log.Info("No new signatory was added")
	} else if len(r["added"]) == 1 {
		log.Info(`==> Successfully added a signatory:`)
	} else {
		log.Info(`==> Successfully added the following signatories:`)
	}

	for i, id := range r["added"] {
		log.Infof(`==> %d. %s`, i+1, id)
	}

	for _, e := range r["errs"] {
		log.Info("Err:", e)
	}

	return nil
}

// RemoveSignatories removes one or more signatories of a cocoon.
func RemoveSignatories(cocoonID string, ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
	defer cc()

	cl := proto_api.NewAPIClient(conn)
	_, err = cl.RemoveSignatories(ctx, &proto_api.RemoveSignatoriesRequest{
		CocoonID: cocoonID,
		IDs:      ids,
	})

	stopSpinner()

	if err != nil {
		return err
	}

	log.Info("Done.")
	return nil
}

// FirewallAllow adds a firewall rule to allow connection to an outgoing destination
func FirewallAllow(dest, port, protocol string) error {
	return nil
}
