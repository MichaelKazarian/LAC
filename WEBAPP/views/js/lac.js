let prevControlMode;
let errorMessage = "";
let infoMessage = "";

let btn9 = document.getElementById("operation9");
btn9.addEventListener('click', async function () {
  let response = await fetch('/radio?id=9',
                            {method: 'GET'});
});

let btn10 = document.getElementById("operation10");
btn10.addEventListener('click', async function () {
  let response = await fetch('/radio?id=10',
                            {method: 'GET'});
});

let btn11 = document.getElementById("operation11");
btn11.addEventListener('click', async function () {
  let response = await fetch('/radio?id=11',
                            {method: 'GET'});
});

// let btnModeManual = document.getElementById("mode-manual");
// btnModeManual.addEventListener('click', async function () {
//   let response = await fetch('/modeset?id=mode-manual',
//                              {method: 'GET'});
// });

let btnModeCycleOnce = document.getElementById("mode-once-cycle");
btnModeCycleOnce.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-once-cycle',
                             {method: 'GET'});
});

let btnModeAuto = document.getElementById("mode-auto");
btnModeAuto.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-auto',
                             {method: 'GET'});
});

let btnStop = document.getElementById("btnStop");
btnStop.addEventListener('click', async function () {
  let response = await fetch('/stop',
                             {method: 'GET'});
});

let circleProgress = document.getElementById("circle-progress");
circleProgress.textFormat = "value";

let stateArea = document.getElementById("state-area");

function getErrorInfo(json) {
  if (json["operationState"].startsWith("error")) return json["operationState"];
  if (json["modeState"].startsWith("error")) return json["modeState"];
  return "";
}

function onCabinetError(error) {
  console.log("** An error occurred during the connection");
  if (errorMessage !== error) {
    stateArea.className = "alert alert-danger";
    stateArea.innerHTML = error;
    errorMessage = error;
  }
};

function setInfoMessage(msg) {
  if (infoMessage !== msg) {
    stateArea.className = "alert alert-info";
    stateArea.innerHTML = msg;
    infoMessage = msg;
  }
}

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

function setDegree(json) {
  circleProgress.value = parseInt(json["degree"]/2);
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

function updModeState(modeId) {
  if (modeId !== prevControlMode) {
    switch (modeId) {
    case "mode-auto":
      btnModeAuto.checked = true;
      btnStop.className = "btn btn-danger btn-lg";
      setOperationsActiveState(false);
      break;
    case "mode-once-cycle":
      btnModeCycleOnce.checked = true;
      setOperationsActiveState(false);
      btnStop.className = "btn btn-danger btn-lg";
      break;
    default:
      // btnModeManual.checked = true;
      btnStop.className = 'btn btn-danger btn-lg invisible';
      setOperationsActiveState(true);
      clearOperationsActiveState();
    }
    prevControlMode = modeId;
  }
}

function updOperationList(json) {
  setOperationState("lRadio1", json["operation1"]);
  setOperationState("lRadio2", json["operation2"]);
  setOperationState("lRadio3", json["operation3"]);
}

async function getCabinetState() {
  //try {
    let response = await fetch("/state");  
    if (response.ok) {
      let json = await response.json();
      let modeId = json["modeId"];
      updModeState(modeId);
      setDegree(json);
      if (modeId !== "mode-manual")
        updOperationList(json);

      let errState = getErrorInfo(json);
      if (errState !== "") {
        onCabinetError(errState);
        updModeState("mode-manual");
      }
      else setInfoMessage(json["modeDescription"]);
    } else {
      onCabinetError("Нема зв'язку з контролером!");
    }
  // } catch (TypeError) {
  //   onCabinetError("Нема зв'язку з контролером!");
  // }
}

function main() {
  function cabinetState() {
        getCabinetState();
        setTimeout(cabinetState, 70);
    }
    setTimeout(cabinetState, 70);
}

if (stateArea !== null) main();
