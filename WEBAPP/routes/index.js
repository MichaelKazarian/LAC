const express = require('express');
const { fork } = require('child_process');

const communicator = fork('routes/communicator.js');
const router = express.Router();

communicator.on("close", (msg) => {
    console.log('Child exited', msg);
});

// use res.render to load up an ejs view file
router.get('/', function(req, res) {  // index page
    res.render('pages/index');
});

router.get('/test', (req, res) => {
    res.send('It works!');
});

let x = "";
router.get('/state', (req, res) => {
    res.send(JSON.parse(x));
});

communicator.on('message', msg => {
    console.log("CHILD MSG", msg);
    x = msg;//JSON.parse(msg);
});

module.exports = router;
