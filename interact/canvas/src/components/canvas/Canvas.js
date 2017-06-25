import React, { Component } from 'react';
import Navigation from '../navigation/Navigation';
import Content from '../content/Content';

// Canvas is the wrapper for all view elements
class Canvas extends Component {
    
    constructor(props){
        super(props);
        this.state = {
            content: "stuff"
        }
    }
    
    setContent(content) {
        this.setState({ content })
    }
    
    render() {
        return <div className="canvas lg">
            <Navigation setContent={this.setContent.bind(this)} /> 
            <Content content={this.state.content} />
        </div>
    }
}

export default Canvas;