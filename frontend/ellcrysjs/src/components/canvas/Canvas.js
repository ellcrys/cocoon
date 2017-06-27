// @flow
import React from 'react';
import Base from '../base/Base'
import Navigation from '../navigation/Navigation';
import {connect} from 'react-redux';
import { Route } from 'react-router'
import {mapStateToProps, matchDispatchToProps} from './mappings'
import type {Manifest, Func} from '../../flow_types/manifest'
import FunctionView from '../function-view/FunctionView';


type State = {
    manifest?: Manifest
}

type PropTypes = {
    updateManifest: (newManifest: {}) => void,
    manifest: Manifest
}

// Canvas is the wrapper for all view elements
class Canvas extends Base {
    state: State
    props: PropTypes
    
    constructor(props: {}){
        super(props)
        this.state = {}
    }
    
    async componentDidMount(): any {
        try {
            let manifest = await this.cocoon.getManifest()
            this.props.updateManifest(manifest)
        } catch(e) {
            console.error("failed to fetch manifest:", e)
        }
    }
    
    /**
     * Create function component
     * 
     * @param {Func} f 
     * @returns 
     * @memberof Canvas
     */
    createFuncComponent(f: Func) {
        return () => { 
            return <FunctionView Func={f} /> 
        }
    } 
    
    /**
     * Create dynamic routes from functions included in the manifest
     * @returns 
     * @memberof Canvas
     */
    createFunctionRoutes() {
        if (!this.props.manifest) return;
        return this.props.manifest.functions.map((f: Func) => {
            return <Route path={"/" + f.name} component={this.createFuncComponent(f)} key={f.name}/>
        })
    }
    
    render() {
        return <div className="canvas lg">
            <Navigation/>
            <div>
                <Route exact path="/" component={FunctionView}/>
                {this.createFunctionRoutes()}
            </div>
        </div>
    }
}

export default connect(mapStateToProps, matchDispatchToProps)(Canvas);