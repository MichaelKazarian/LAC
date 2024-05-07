const express = require('express');
const { fork } = require('child_process');

const communicator = fork('controller/process_controller.js');
const router = express.Router();

/*    Queries serving     */

// use res.render to load up an ejs view file
router.get('/', function(req, res) {  // index page
    res.render('pages/index');
});

router.get('/test', (req, res) => {
    res.send('It works!');
});

router.get('/state', (req, res) => {
    // console.log("JSON", result);
    res.send(dataInput);
});

router.get("/radio", (req, res) => {
    communicator.send("radio&r"+req.query.id);
    res.send("ok");
});

/*     Communications with equipment     */

let dataInput = [];
communicator.on('message', msg => {
    dataInput = parseInput(msg);//JSON.parse(msg);
});

communicator.on("close", (msg) => {
    console.log('Child exited', msg);
});

function parseInput(data) {
    jsonObj = JSON.parse(data);
    result = {
        degree: jsonObj["degree"],
        error: jsonObj["error"],
        state: "ok"
    };
    a = 1;
    for (var i in jsonObj["rawinput"]) {
        result[`param${a}`] = jsonObj["rawinput"][i];
        a++;
    }
    return result;
}

module.exports = router;
