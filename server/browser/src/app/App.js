import React from 'react';
import { Route, Switch } from "react-router-dom"
import Dashboard from './browser/Dashboard'
import Login from './browser/Login'



function App() {
  return (
    <Switch>
      <Route path={"/login"} component={Login} />
      <Route path={"/:dashboard?/*"} component={Dashboard} />
    </Switch>
  );
}

export default App;
