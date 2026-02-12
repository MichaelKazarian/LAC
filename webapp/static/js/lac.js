let isConnected = true;
let prevControlMode;
let isOperationsRendered = false;
let isPausedGlobal = false;
let lastPausedState = null;
let errorMessage = "";
let infoMessage = "";
let warningMessage = "";
let manualOperations = [];
let lastActiveId = null;
let lastError = "";

let btnManual = document.getElementById("mode-manual");
btnManual.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-manual',
                             {method: 'GET'});
});

let lbModeCycleOnce = document.getElementById("lb-mode-once-cycle");
let btnModeCycleOnce = document.getElementById("mode-once-cycle");
btnModeCycleOnce.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-once-cycle',
                             {method: 'GET'});
});

let lbModeAuto = document.getElementById("lb-mode-auto");
let btnModeAuto = document.getElementById("mode-auto");
btnModeAuto.addEventListener('click', async function () {
  let response = await fetch('/modeset?id=mode-auto',
                             {method: 'GET'});
});

let btnPause = document.getElementById("btnPause");
btnPause.addEventListener('click', async function () {
  let targetState = !isPausedGlobal; 
  setInfoMessage(`Sending pause command: ${targetState}`);
  let response = await fetch(`/pause?set=${targetState}`, { method: 'GET' });
  if (!response.ok) {
    onCabinetError("Помилка відправки команди паузи");
  }
});

let btnSafety = document.getElementById("btnSafety");
btnSafety.addEventListener('click', async function () {
  await fetch('/safety', { method: 'GET' });
});

let circleProgress = document.getElementById("circle-progress");
circleProgress.textFormat = "value";

let stateArea = document.getElementById("state-area");
let counterContainer = document.getElementById("container-counter");
let productCounter = document.getElementById("product-counter");
let lbProductCounter = document.getElementById("lb-product-counter");

function arraysEqual(a1,a2) {
  /* WARNING: arrays must not contain {objects} or behavior may be undefined */
  return JSON.stringify(a1)==JSON.stringify(a2);
}

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

function setWarningMessage(msg) {
  if (warningMessage !== msg) {
    stateArea.className = "alert alert-warning";
    stateArea.innerHTML = `${infoMessage}, ${msg}`;
    warningMessage = msg;
  }
}

function renderOperations(operations) {
  const leftCol = document.getElementById("ops-col-left");
  const rightCol = document.getElementById("ops-col-right");
  
  if (!leftCol || !rightCol || !operations) return;

  // Розраховуємо точку поділу (якщо 19, то limit = 10)
  const limit = Math.ceil(operations.length / 2);
  
  let leftHtml = "";
  let rightHtml = "";

  operations.forEach((op, index) => {
    const id = op[0];   // Ключ/ID операції
    const name = op[1]; // Те, що в Go ми прописали як UserName
    
    const html = `
      <div class="row mx-1 my-1">
                <input type="radio" class="btn-check" name="radio-operation" id="${id}" autocomplete="off">
                <label class="btn btn-outline-primary btn-lg" for="${id}" id="lRadio_${id}">
                  ${name}
                </label>
      </div>`;

    if (index < limit) {
      leftHtml += html;
    } else {
      rightHtml += html;
    }
  });

  // Оновлюємо DOM
  leftCol.innerHTML = leftHtml;
  rightCol.innerHTML = rightHtml;

  // Прив'язка кліків
  operations.forEach((op) => {
    const id = op[0];
    const el = document.getElementById(id);
    if (el) {
      el.onclick = () => invokeOperation(id);
    }
  });
}

async function invokeOperation(id) {
  console.log(`Sending operation: ${id}`);
  let response = await fetch(`/radio?id=${id}`, { method: 'GET' });
  if (!response.ok) {
    onCabinetError(`Помилка виконання: ${id}`);
  }
}

function setOperationState(elementId, value) {
  var element = document.getElementById(elementId);
  if (!element) return;
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
  if (productCounter) {
    productCounter.innerHTML = json["counter"] || 0;
  }
}

function setOperationsActiveState(state) {
  let operations = document.getElementsByName("radio-operation");
  for (let i=0; i<operations.length; i++) {
    let r = operations[i];
    r.disabled = !state;
  }
}

function clearOperationsActiveState() {
  let operations = document.getElementsByName("radio-operation");
  operations.forEach(r => {
    r.checked = false;
    let l = document.querySelector(`label[for="${r.id}"]`);
    if (l) {
      l.className = "btn btn-outline-secondary btn-lg";
    }
  });
}

function updAvailableManualOperations(json) {
  const newOps = json["manualOperations"];
  // Якщо список дозволених операцій не змінився — виходимо
  if (arraysEqual(manualOperations, newOps)) return;

  // 1. Скидаємо всі кнопки операцій до "вимкненого" стану (сірі)
  let allRadios = document.getElementsByName("radio-operation");
  allRadios.forEach(r => {
    r.disabled = true;
    let l = document.querySelector(`label[for="${r.id}"]`);
    if (l) {
      // Робимо кнопку сірою
      l.className = "btn btn-outline-secondary btn-lg";
    }
  });

  // 2. Вмикаємо лише ті, що дозволені бекендом за кутом/станом
  newOps.forEach(opId => {
    const radio = document.getElementById(opId);
    const label = document.querySelector(`label[for="${opId}"]`);
    if (radio && label) {
      radio.disabled = false;
      // Робимо кнопку синьою та жирною
      label.className = "btn btn-outline-primary btn-lg fw-bold";
    }
  });

  manualOperations = newOps;
}

function updPauseButton(isPaused, modeId) {
  let btnPause = document.getElementById("btnPause");
  if (isPaused === lastPausedState || modeId === "mode-manual") return;
  if (isPaused) {
    btnPause.innerHTML = "ПРОДОВЖИТИ";
    btnPause.className = "btn btn-success btn-lg blink"; // blink можна додати в CSS для уваги
    btnManual.disabled = false;
    btnManual.classList.remove("btn-outline-secondary");
    btnManual.classList.add("btn-secondary", "fw-bold");
  } else {
    btnPause.innerHTML = "ПАУЗА";
    btnPause.className = "btn btn-warning btn-lg";
    btnManual.disabled = true;
    btnManual.classList.add("btn-outline-secondary");
    btnManual.classList.remove("btn-secondary", "fw-bold");
  }
  lastPausedState = isPaused;
}

function updModeState(modeId) {
  if (modeId !== prevControlMode) {
    errorMessage = "";
    infoMessage = "";
    warningMessage = "";

    switch (modeId) {
      case "mode-auto":
      case "mode-once-cycle":
        const isAuto = (modeId === "mode-auto");
        btnModeAuto.checked = isAuto;
        btnModeCycleOnce.checked = !isAuto;
        btnPause.classList.remove("invisible");
        lbModeAuto.classList.add("invisible");
        lbModeCycleOnce.classList.add("invisible");
        if (isAuto) {
          counterContainer.classList.remove("invisible");
        } else {
          counterContainer.classList.add("invisible");
        }
        setOperationsActiveState(false);
        break;
      default:
        btnManual.checked = true;
        btnManual.disabled = false;
        btnPause.classList.add("invisible");
        lbModeAuto.className = "btn btn-outline-success btn-lg";
        lbModeCycleOnce.className = "btn btn-outline-success btn-lg";
        counterContainer.classList.add("invisible");
        setOperationsActiveState(true);
        clearOperationsActiveState();
    }
    prevControlMode = modeId;
  }
}

async function getCabinetState() {
  try {
    let response = await fetch("/state");  
    if (response.ok) {
      let json = await response.json();
      if (!isConnected) {
        // TODO в окрему функцію
        isConnected = true;
        errorMessage = ""; // Скидаємо стару помилку зв'язку
        infoMessage = "";
        warningMessage = "";
        console.log("Зв'язок відновлено");
      }
      // ПЕРЕВІРКА ТА РЕНДЕР ОПЕРАЦІЙ (Тільки один раз!)
      if (!isOperationsRendered && json["OperationsList"]) {
        renderOperations(json["OperationsList"]);
        isOperationsRendered = true;
      }
      let modeId = json["modeId"];
      isPausedGlobal = json["isPaused"];
      /* console.log("modeId "+modeId+" isPausedGlobal "+isPausedGlobal) */
      updSafetyButton(json);
      updPauseButton(isPausedGlobal, modeId);
      updModeState(modeId);
      setDegree(json);
      if (modeId === "mode-manual") {
        updAvailableManualOperations(json);
      } else {
        setOperationsActiveState(false);
        manualOperations = []; 
      }
      updActiveOperation(json["ActiveOperation"]);
      let errState = getErrorInfo(json);
      if (errState !== "") {
        onCabinetError(errState);
        updModeState("mode-manual");
      }
      else if (json["modeState"].startsWith("warning-")) setWarningMessage(json["modeState"].split("-")[1]);
      else setInfoMessage(json["modeDescription"]);
    } else {
      isConnected = false;
      onCabinetError(`Помилка сервера: ${response.status}`);
    }
  } catch (TypeError) {
    if (isConnected) { // Фізична відсутність зв'язку (Network Error)
      isConnected = false;
      onCabinetError("Зв'язок з контролером відсутній!");
      setOperationsActiveState(false); // Блокуємо все від гріха подалі
    }
  }
}

function updActiveOperation(activeId) {
  // Якщо ID не змінився, нічого не робимо — виходимо миттєво
  if (activeId === lastActiveId) return;
  console.log(`🔄 Зміна активної операції: ${lastActiveId} -> ${activeId}`);
  // 1. Скидаємо попередню активну кнопку (якщо вона була)
  if (lastActiveId) {
    let lastLabel = document.querySelector(`label[for="${lastActiveId}"]`);
    if (lastLabel) {
      lastLabel.classList.remove("btn-warning", "blink");
      lastLabel.classList.add("btn-outline-primary");
    }
    // Тут не було скидання r.checked = false, тому кнопка залишалася "вибраною"
  }

  // 2. Підсвічуємо нову активну кнопку
  if (activeId && activeId !== "") {
    let currentRadio = document.getElementById(activeId);
    let currentLabel = document.querySelector(`label[for="${activeId}"]`);
    if (currentLabel) {
      currentLabel.classList.remove("btn-outline-primary", "btn-outline-secondary");
      currentLabel.classList.add("btn-warning", "fw-bold", "blink");
      if (currentRadio) currentRadio.checked = true;
    }
  }
  lastActiveId = activeId; // Запам'ятовуємо новий стан
}

// ВАРІАНТ з "відтисканям" кнопки
/* function updActiveOperation(activeId) {
 *   if (activeId === lastActiveId) return;
 * 
 *   console.log(`🔄 Стан операцій змінено: ${lastActiveId} -> ${activeId}`);
 * 
 *   // 1. Скидаємо попередню активну кнопку
 *   if (lastActiveId) {
 *     let lastRadio = document.getElementById(lastActiveId);
 *     let lastLabel = document.querySelector(`label[for="${lastActiveId}"]`);
 *     
 *     if (lastLabel) {
 *       lastLabel.classList.remove("btn-warning", "blink");
 *       // ПЕРЕКОНУЄМОСЯ, що вона повертається до правильного кольору
 *       if (manualOperations.includes(lastActiveId)) {
 *         lastLabel.className = "btn btn-outline-primary btn-lg fw-bold";
 *       } else {
 *         lastLabel.className = "btn btn-outline-secondary btn-lg";
 *       }
 *     }
 *     // ВАЖЛИВО: "відтискаємо" кнопку
 *     if (lastRadio) lastRadio.checked = false;
 *   }
 * 
 *   // 2. Підсвічуємо нову активну кнопку (якщо вона є)
 *   if (activeId && activeId !== "") {
 *     let currentRadio = document.getElementById(activeId);
 *     let currentLabel = document.querySelector(`label[for="${activeId}"]`);
 *     if (currentLabel) {
 *       currentLabel.classList.remove("btn-outline-primary", "btn-outline-secondary", "btn-outline-secondary");
 *       currentLabel.classList.add("btn-warning", "fw-bold", "blink");
 *       if (currentRadio) currentRadio.checked = true;
 *     }
 *   }
 * 
 *   lastActiveId = activeId;
 * } */

function updSafetyButton(json) {
  const btn = document.getElementById("btnSafety");
  const isLocked = json["isLocked"];
  const activeOp = json["ActiveOperation"];
  const currentError = json["stopReason"];
  const mode = json["modeId"];

  // Якщо система заблокована і з'явилося повідомлення, яке ми ще не показували
  if (isLocked && currentError && currentError !== lastError) {
    lastError = currentError; 
    // Викликаємо вашу функцію обробки помилки
    if (typeof onCabinetError === "function") {
      onCabinetError(currentError);
    }
  }
  // Скидаємо збережену помилку, коли систему розблоковано
  if (!isLocked) {
    lastError = "";
  }

  if (isLocked) {
    btn.innerHTML = "РОЗБЛОКУВАТИ";
    btn.className = "btn btn-success btn-lg w-100 shadow-sm";
    hideRightPanel()
  } else {
    btn.innerHTML = "СТОП";
    btn.className = "btn btn-danger btn-lg w-100 shadow-sm";
    showRightPanel()
  }
  // Показуємо кнопку лише якщо:
  // 1. Система заблокована СТОПом
  // 2. Йде будь-яка операція (ActiveOperation не пуста)
  // 3. Ми НЕ в ручному режимі (тобто в Авто або Одиночному)
  if (isLocked || activeOp !== "" || mode !== "mode-manual") {
    btn.classList.remove("invisible");
  } else {
    btn.classList.add("invisible");
  }
}

function hideRightPanel() {
  document.getElementById("rightPanel").classList.add("d-none");
}

function showRightPanel() {
  document.getElementById("rightPanel").classList.remove("d-none");
}

function main() {
  function cabinetState() {
    getCabinetState();
    setTimeout(cabinetState, 70);
  }
  setTimeout(cabinetState, 70);
}

if (stateArea !== null) main();
