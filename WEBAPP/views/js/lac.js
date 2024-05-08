let STATE_UPD_TIMEOUT = 0;

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

function updDashboardElements(object) {
    document.getElementById("state-area").innerHTML = object["state"];
    setRadioState("lRadio1", object["param1"]);
    setRadioState("lRadio2", object["param2"]);
    setRadioState("lRadio3", object["param3"]);
  // document.getElementById("network_state").innerHTML = "";
  // STATE_UPD_TIMEOUT = 0;
  /* console.log(document.getElementById("my_smoke").className); */
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

