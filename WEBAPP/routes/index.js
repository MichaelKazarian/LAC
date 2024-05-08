const express = require('express');
const { fork } = require('child_process');

const communicator = fork('controller/process_controller.js');
const router = express.Router();

/*    Queries serving     */

// use res.render to load up an ejs view file
router.get('/', function(req, res) {  // index page
  const params = {
    "modes": [
      {"id": "mode-manual", name: "Вручну", class: "btn-outline-warning"},
      {"id": "mode-once-cycle", name: "Одиночний цикл", class: "btn-outline-success"},
      {"id": "mode-auto", name: "Автомат", class: "btn-outline-success"}
    ]};
  res.render('pages/index', params);
});

router.get('/state', (req, res) => {
    res.send(dataInput);
});

router.get("/radio", (req, res) => {
  communicator.send("radio&r"+req.query.id);
  res.send("ok");
});

router.get("/modeset", (req, res) => {
  communicator.send(req.query.id);
  res.send("ok");
});


/*     Communications with equipment     */

let dataInput = {
  degree: undefined,
  error: undefined,
  modeId: 'mode-manual',
  modeDescription: 'Manual mode',
  modeStatus: 'error',
  state: undefined,
};

communicator.on('message', msg => {
  if (!msg.includes("type")) {
    console.log(msg);
    return;
  };
  json = JSON.parse(msg);
  let t = json["type"];
  if (t === "input")  setInputData(json);
  if (t === "mode") setMode(json);
  // console.log(dataInput);
});

communicator.on("close", (msg) => {
    console.log('Child exited', msg);
});

function setMode(json) {
  dataInput["modeId"] = json["modeId"];
  dataInput["modeDescription"] = json["modeDescription"];
  dataInput["modeStatus"] = json["modeStatus"];
}

function setInputData(json) {
  dataInput["degree"] = json["degree"];
  dataInput["error"] = json["error"];
  dataInput["state"] = "ok";
  a = 1;
  for (var i in json["rawinput"]) {
    dataInput[`param${a}`] = json["rawinput"][i];
    a++;
  }
}

module.exports = router;
