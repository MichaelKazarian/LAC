const express = require('express');
const app = express();
const path = require('path');
const routes = require('./routes/index');

app.use("/css",express.static(path.join(__dirname, "node_modules/bootstrap/dist/css")));
app.use("/css",express.static(path.join(__dirname, "views/css")));
app.use("/js",express.static(path.join(__dirname, "node_modules/bootstrap/dist/js")));
app.use("/js",express.static(path.join(__dirname, "views/js")));
app.set('view engine', 'ejs'); // set the view engine to ejs

app.use('/', routes);

module.exports = app;
