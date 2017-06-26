import React from 'react';
import ReactDOM from 'react-dom';
import App from './components/app';
import {Provider} from 'react-redux'
import './dist/canvas.css';
import {createStore} from 'redux';
import reducers from './reducers'
const store = createStore(reducers)

ReactDOM.render(
    <Provider store={store}>
        <App />
    </Provider> , 
    document.getElementById('root'));