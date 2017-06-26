// @flow
import React from 'react';
import Base from '../base/Base'
import Navigation from '../navigation/Navigation';
import Content from '../content/Content';
import {connect} from 'react-redux';
import {mapStateToProps, matchDispatchToProps} from './mappings'


type State = {
    manifest?: {}
}

type PropTypes = {
    updateManifest: (newManifest: {}) => void
}

// Canvas is the wrapper for all view elements
class Canvas extends Base {
    state: State
    propTypes: PropTypes
    
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
    
    render() {
        return <div className="canvas lg">
            <Navigation /> 
            <Content />
        </div>
    }
}

export default connect(mapStateToProps, matchDispatchToProps)(Canvas);