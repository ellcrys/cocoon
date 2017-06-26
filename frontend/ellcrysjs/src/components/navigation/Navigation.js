// @flow
import React from 'react'
import Base from '../base/Base'
import {connect} from 'react-redux'
import {mapStateToProps, matchDispatchToProps} from './mappings'

type PropTypes = {
}

type State = {
}

class Navigation extends Base {
    state: State
    propTypes: PropTypes
    
    constructor(props: PropTypes){
        super(props)
        this.state = {}
    }
    
    render() {
        return <div className="navigation">
            
        </div>
    }
}


export default connect(mapStateToProps, matchDispatchToProps)(Navigation);