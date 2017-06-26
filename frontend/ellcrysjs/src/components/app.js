import 'whatwg-fetch';
import Promise from 'promise-polyfill'; 
import React, { Component } from 'react';
import Canvas from './canvas/Canvas';

if (!window.Promise) {
	window.Promise = Promise;
}

class App extends Component {
	render() {
		return (
			<Canvas />
		);
	}
}

export default App;
