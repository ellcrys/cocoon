import React, { Component } from 'react';
import CardML from '../../cardml/cardml'

class Navigation extends Component {
    
    testFetch() {
        const cml = new CardML()
        var component = cml.parse(`
        <card style="height: 100px; border: #000 1px solid;" direction="row" wrap="nowrap" justify-content="center" align-items="center" align-content="center">
            <card order="2" grow="1" align-self="flex-end">hello</card>
            <card order="1">world</card>
        </card>`) 
        this.props.setContent(component)
    }
    
    render() {
        return <div onClick={this.testFetch.bind(this)} className="navigation">
            <b>Navigation</b> 
            <a>Load</a>
        </div>
    }
}

export default Navigation;