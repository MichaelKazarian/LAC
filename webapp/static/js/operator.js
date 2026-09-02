// static/js/operator.js

export let isConnected = true;
export let prevControlMode = null;
export let isPausedGlobal = false;
export let lastPausedState = null;
export let lastSafetyVisibility = null;
export let lastSafetyLocked = null;

export const stateArea = document.getElementById("state-area");
export const productCounter = document.getElementById("product-counter");
export const btnModeCycleOnce = document.getElementById("mode-once-cycle");
export const btnModeAuto = document.getElementById("mode-auto");
export const btnModeManual = document.getElementById("mode-manual"); // Присутній на сторінці наладчика
export const btnPause = document.getElementById("btnPause");
export const btnSafety = document.getElementById("btnSafety");

// Ініціалізація базових подій
export function initBaseEvents() {
  if (btnModeCycleOnce) {
    btnModeCycleOnce.addEventListener('click', () => fetch('/modeset?id=mode-once-cycle'));
  }
  if (btnModeAuto) {
    btnModeAuto.addEventListener('click', () => fetch('/modeset?id=mode-auto'));
  }
  if (btnModeManual) {
    btnModeManual.addEventListener('click', () => fetch('/modeset?id=mode-manual'));
  }

  if (btnPause) {
    btnPause.addEventListener('click', async () => {
      const targetState = !isPausedGlobal;
      const response = await fetch(`/pause?set=${targetState}`);
      if (!response.ok) setMessage("error", "Помилка відправки команды паузи");
    });
  }

  if (btnSafety) {
    btnSafety.addEventListener('click', () => fetch('/safety'));
  }
}

export function setMessage(type, text) {
  if (!stateArea) return;
  stateArea.innerHTML = text;
  switch (type) {
    case "error":
      stateArea.className = "alert alert-danger text-center fs-4 p-4 my-2 border";
      break;
    case "warning":
      stateArea.className = "alert alert-warning text-center fs-4 p-4 my-2 border";
      break;
    case "info":
      stateArea.className = "alert alert-info text-center fs-4 p-4 my-2 border";
      break;
    default:
      stateArea.className = "alert alert-light text-center fs-4 p-4 my-2 border";
  }
}

export function updPauseButton(isPaused) {
  if (!btnPause || isPaused === lastPausedState) return;
  if (isPaused) {
    btnPause.innerHTML = "ПРОДОВЖИТИ";
    btnPause.className = "btn btn-success btn-lg w-100 mb-3 py-2 blink";
  } else {
    btnPause.innerHTML = "ПАУЗА";
    btnPause.className = "btn btn-warning btn-lg w-100 mb-3 py-2";
  }
  lastPausedState = isPaused;
}

export function updModeState(modeId) {
  if (modeId !== prevControlMode) {
    const isAuto = (modeId === "mode-auto");
    const isSingle = (modeId === "mode-once-cycle");
    const isManual = (modeId === "mode-manual");

    if (btnModeAuto) btnModeAuto.checked = isAuto;
    if (btnModeCycleOnce) btnModeCycleOnce.checked = isSingle;
    if (btnModeManual) btnModeManual.checked = isManual;

    if (isManual) {
      if (btnPause) btnPause.classList.add("invisible");
    } else {
      if (btnPause) btnPause.classList.remove("invisible");
    }

    prevControlMode = modeId;
  }
}

export function updateSafetyIfNeeded(json) {
  const isLocked = json["isLocked"];
  const mode = json["modeId"];
  const shouldBeVisible = (isLocked || mode !== "mode-manual");

  if (lastSafetyVisibility === shouldBeVisible && lastSafetyLocked === isLocked) return;

  lastSafetyVisibility = shouldBeVisible;
  lastSafetyLocked = isLocked;

  if (isLocked) {
    btnSafety.innerHTML = "РОЗБЛОКУВАТИ";
    btnSafety.className = "btn btn-success btn-lg w-100 shadow-sm";
    document.getElementById("rightPanel")?.classList.add("d-none");
  } else {
    btnSafety.innerHTML = "СТОП";
    btnSafety.className = "btn btn-danger btn-lg w-100 shadow-sm";
    document.getElementById("rightPanel")?.classList.remove("d-none");
  }

  btnSafety.classList.toggle("invisible", !shouldBeVisible);
}

export function getErrorInfo(json) {
  if (json["operationState"] && json["operationState"].startsWith("error")) return json["operationState"];
  if (json["modeState"] && json["modeState"].startsWith("error")) return json["modeState"];
  if (json["isLocked"]) return json["stopReason"];
  return "";
}

// Функція оновлення базового UI
export function applyBaseState(json) {
  isPausedGlobal = json["isPaused"];
  const modeId = json["modeId"];

  updateSafetyIfNeeded(json);
  updPauseButton(isPausedGlobal);
  updModeState(modeId);

  if (productCounter) {
    productCounter.innerHTML = json["counter"] || 0;
  }

  const errState = getErrorInfo(json);
  if (errState !== "") {
    setMessage("error", errState);
  } else if (json["modeState"] && json["modeState"].startsWith("warning-")) {
    setMessage("warning", json["modeState"].split("-")[1]);
  } else {
    setMessage("info", json["modeDescription"]);
  }
}

// Автономний запуск для сторінки оператора
if (document.body.dataset.scriptRole === "operator") {
  document.addEventListener("DOMContentLoaded", () => {
    initBaseEvents();
    
    async function loop() {
      try {
        const response = await fetch("/state");
        if (response.ok) {
          const json = await response.json();
          applyBaseState(json);
        }
      } catch (e) {
        setMessage("error", "Втрачено зв'язок з контролером!");
      }
      setTimeout(loop, 100);
    }
    loop();
  });
}

// Загальні системні вікна
window.openStatusWindow = function() {
  window.open('/status', 'StatusPage');
};

window.confirmShutdown = function() {
  const myModal = new bootstrap.Modal(document.getElementById('shutdownModal'));
  myModal.show();
};

window.doShutdown = function() {
  fetch('/shutdown').then(r => {
    if (r.ok) {
      document.body.innerHTML = `<div class="d-flex justify-content-center align-items-center vh-100 bg-dark text-white"><h1>СИСТЕМА ВИМКНЕНА</h1></div>`;
    }
  });
};
