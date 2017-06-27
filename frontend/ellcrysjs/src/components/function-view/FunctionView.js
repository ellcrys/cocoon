// @flow
import React from 'react';
import Base from '../base/Base'
import type {Manifest, Func } from '../../flow_types/manifest'
import {InvokeError} from '../../errors/errors'
import type {View} from '../../flow_types/view'

type State = {
    manifest?: Manifest,
    content?: string
}

type PropTypes = {
    Func: Func
}

/**
 * FunctionView defines a component for displaying 
 * a card from a contract function
 * 
 * @class FunctionView
 * @extends {Base}
 */
class FunctionView extends Base {
    state: State
    props: PropTypes
    
    constructor(props: {}) {
        super(props)
        this.state = {}
    }
    
    /**
     * Fetches the function's view by sending an invoke request to 
     * retrieve a valid view. The view is then passed to the EML parser 
     * for processing and conversion to a react view component
     * 
     * @memberof FunctionView
     */
    fetchView() {
        this.cocoon.getView(this.props.Func.name).then((view: View) => {
            try {
                let m = view.markup || ""
                this.eml.parse(m)
                this.setState({ content: this.eml.parse(m) })
            } catch(e) {
                console.error("failed to parse view:", e)
            }
        }).catch((err) => {
            console.error("failed to get view:", (err: InvokeError).message)     // TODO: should we be logging invoke errors?
        })
    }

    componentWillMount() {
        this.fetchView()
    }

    render() {
        return <div className="content">
            {this.state.content}
        </div>
    }
}

export default FunctionView;