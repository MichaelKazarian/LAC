let STATE_UPD_TIMEOUT = 0;
let prevControlMode;

function setOperationState(elementId, value) {
  var element = document.getElementById(elementId);
  var currentClass = element.className;
  if (value === 0) {
    element.className = "btn btn-primary btn-lg";
    clearOperationsActiveState();
    element.checked = true;
  } else if (value === 1) {
    element.className = "btn btn-outline-primary btn-lg";
  } else {
    element.className = "btn btn-danger btn-lg";
  }
}  

function onCabinetError() {
  console.log("** An error occurred during the transaction");
  STATE_UPD_TIMEOUT = 200;
  document.getElementById("state-area").innerHTML = "Нема зв'язку";
};

if (document.getElementById("state-area")) {
    function cabinetState() {
        getCabinetState();
        setTimeout(cabinetState, 70);
    }
    setTimeout(cabinetState, 70);
}

btn9 = document.getElementById("operation9");
btn9.addEventListener('click', async function () {
  console.log("UOD9");
  let response = await fetch('/radio?id=9',
                            {method: 'GET'});
});

btn10 = document.getElementById("operation10");
btn10.addEventListener('click', async function () {
  let response = await fetch('/radio?id=10',
                            {method: 'GET'});
});

btn11 = document.getElementById("operation11");
btn11.addEventListener('click', async function () {
  let response = await fetch('/radio?id=11',
                            {method: 'GET'});
});

btnModeManual = document.getElementById("mode-manual");
btnModeManual.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-manual',
                             {method: 'GET'});
});

btnModeCycleOnce = document.getElementById("mode-once-cycle");
btnModeCycleOnce.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-once-cycle',
                             {method: 'GET'});
});

btnModeAuto = document.getElementById("mode-auto");
btnModeAuto.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-auto',
                             {method: 'GET'});
});

let circleProgress = document.getElementById("circle-progress");
circleProgress.textFormat = "value";

function setDegree(json) {
  circleProgress.value = parseInt(json["degree"])/2;
}

function setOperationsActiveState(state) {
  let operations = document.getElementsByName("radio-operation");
  for (let i=0; i<operations.length; i++) {
    let r = operations[i];
    r.disabled = !state;
  }
  if (!state) {
    clearOperationsActiveState();
  }
}

function clearOperationsActiveState() {
  let operations = document.getElementsByName("radio-operation");
  for (let i=0; i<operations.length; i++) {
    let r = operations[i];
    r.checked = false;
  }
}


function updModeState(object) {
  if (object["modeId"] !== prevControlMode) {
    switch (object["modeId"]) {
    case "mode-auto":
      btnModeAuto.checked = true;
      setOperationsActiveState(false);
      break;
    case "mode-once-cycle":
      btnModeCycleOnce.checked = true;
      setOperationsActiveState(false);
      break;
    default:
      btnModeManual.checked = true;
      setOperationsActiveState(true);
      clearOperationsActiveState();
    }
    prevControlMode = object["modeId"];
  }
}

function updOperationList(object) {
  setOperationState("lRadio1", object["operation1"]);
  setOperationState("lRadio2", object["operation2"]);
  setOperationState("lRadio3", object["operation3"]);
}

async function getCabinetState() {
    let response = await fetch("/state");
    if (response.ok) {
        let json = await response.json();
      updModeState(json);
      setDegree(json);
      if (json["modeId"] !== "mode-manual")
        updOperationList(json);
    } else {
        onCabinetError();
    }
}
