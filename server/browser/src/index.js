import React from 'react';
import ReactDOM from 'react-dom';
import { Router } from "react-router-dom"
import { Provider } from "react-redux"

//import 'bootstrap/dist/css/bootstrap.min.css';
import './bootstrap.css';
import './index.css';
import './dark.scss';

import App from './app/App';
import * as serviceWorker from './serviceWorker';
import { library } from '@fortawesome/fontawesome-svg-core'
import { faTrash, faSignInAlt, faCircle, faCog, faPowerOff, faFileCode, faUpload } from '@fortawesome/free-solid-svg-icons'

import history from "./app/history"
import configureStore from "./app/store/configure-store"

library.add(faTrash, faSignInAlt, faCircle, faCog, faPowerOff, faFileCode, faUpload )

const store = configureStore()

ReactDOM.render(
    <Provider store={store}>
      <Router history={history}>
        <App />
      </Router>
    </Provider>,
    document.getElementById("root")
)

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
