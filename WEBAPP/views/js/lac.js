let STATE_UPD_TIMEOUT = 0;
let prevControlMode;

function setRadioState(elementId, value) {
  var NAME = document.getElementById(elementId);
  var currentClass = NAME.className;
  if (value === 0) {
    NAME.className = "btn btn-success btn-lg";
  } else if (value === 1) {
    NAME.className = "btn btn-outline-secondary btn-lg";
  } else {
    NAME.className = "btn btn-danger btn-lg";
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

btn9 = document.getElementById("radio9");
btn9.addEventListener('click', async function () {
  let response = await fetch('/radio?id=9',
                            {method: 'GET'});
});

btn10 = document.getElementById("radio10");
btn10.addEventListener('click', async function () {
  let response = await fetch('/radio?id=10',
                            {method: 'GET'});
});

btn11 = document.getElementById("radio11");
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

function setOperationsActive(state) {
  let operations = document.getElementsByName("radio-operation");
  for (let i=0; i<operations.length; i++) {
    let r = operations[i];
    r.disabled = !state;
  }
}

function updDashboardElements(object) {
  if (object["modeId"] !== prevControlMode) {
    switch (object["modeId"]) {
    case "mode-auto":
      btnModeAuto.checked = true;
      setOperationsActive(false);
      break;
    case "mode-once-cycle":
      btnModeCycleOnce.checked = true;
      setOperationsActive(false);
      break;
    default:
      btnModeManual.checked = true;
      setOperationsActive(true);
    }
    prevControlMode = object["modeId"];
  }
}

function updControlState(object) {
}



async function getCabinetState() {
    let response = await fetch("/state");
    if (response.ok) {
        let json = await response.json();
        updDashboardElements(json);
    } else {
        onCabinetError();
    }
}
