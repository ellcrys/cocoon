import React from 'react';
import ReactDOM from 'react-dom';
import App from './components/app';
import {Provider} from 'react-redux'
import './dist/canvas.css';
import {createStore, applyMiddleware} from 'redux';
import reducers from './reducers'
import createHistory from 'history/createBrowserHistory'
import { ConnectedRouter, routerReducer, routerMiddleware } from 'react-router-redux'

// Create a history 
const history = createHistory()

// Build the middleware for intercepting and dispatching navigation actions
const middleware = routerMiddleware(history)

// Create store and apply router middleware
reducers.router = routerReducer
const store = createStore(reducers, applyMiddleware(middleware))

ReactDOM.render(
    <Provider store={store}>
        <ConnectedRouter history={history}>
            <App />
        </ConnectedRouter>
    </Provider> , 
    document.getElementById('root'));