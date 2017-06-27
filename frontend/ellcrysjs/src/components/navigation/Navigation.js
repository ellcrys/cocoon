// @flow
import React, {Element} from 'react'
import Base from '../base/Base'
import { connect } from 'react-redux'
import { mapStateToProps, matchDispatchToProps } from './mappings'
import type {Manifest, Func} from '../../flow_types/manifest'

export type PropTypes = {
    manifest: Manifest
}

class Navigation extends Base {
    propTypes: PropTypes

    /**
     * List all available functions included in the manifest
     * 
     * @returns 
     * @memberof Navigation
     */
    listFunctions(): Array<Element>|null {
        if (!this.props.manifest) {
            return null
        }
        return this.props.manifest.functions.map((func: Func) => {
            return <a href={"/" + func.name} className="nav-item" key={func.name}>{func.title}</a>
        })
    }

    render() {
        return <div className="navigation">
            <nav className="nav">
                <div className="nav-left">
                    <div className="nav-item">Contract App</div>
                </div>  
                <div className="nav-right nav-menu">
                    {this.listFunctions()}
                </div>
            </nav>
        </div>
    }
}


export default connect(mapStateToProps, matchDispatchToProps)(Navigation);