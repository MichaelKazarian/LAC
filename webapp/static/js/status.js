// status.js
let ioLabels = { in: {}, out: {} };

async function loadIOMap() {
  try {
    const response = await fetch('/api/io-map');
    const data = await response.json();
    ioLabels.in = data.in || {};
    ioLabels.out = data.out || {};
    console.log("IO Map loaded:", ioLabels);
  } catch (err) {
    console.error("Failed to load IO map:", err);
  }
}

// Головна функція ініціалізації
async function init() {
  await loadIOMap();      // 1. Завантажуємо імена пінів один раз
  updateStatus();         // 2. Перший рендер даних
  setInterval(updateStatus, 250); // 3. Запускаємо цикл
}

function returnToMain() {
  if (window.opener && !window.opener.closed) {
    window.opener.focus();   // переключитися на основну вкладку
    window.close();          // закрити status
  } else {
    // якщо головна вкладка вже закрита
    window.location.href = "/";
  }
}

function updateStatusLabel(id, isOnline) {
  const el = document.getElementById(id);
  if (isOnline) {
    el.innerText = "ONLINE";
    el.classList.add('bg-light', 'text-success');
    el.classList.remove('bg-danger', 'text-white');
  } else {
    el.innerText = "OFFLINE";
    el.classList.add('bg-danger', 'text-white');
    el.classList.remove('bg-light', 'text-success');
  }
}

function renderGrid(containerId, dataArray, dotClassPrefix, colorClass) {
  const grid = document.getElementById(containerId);
  
  // Визначаємо, яку мапу імен використовувати (in чи out)
  const currentLabels = containerId.includes('inputs') ? ioLabels.in : ioLabels.out;

  if (grid.children.length === 0) {
    dataArray.forEach((_, i) => {
      const div = document.createElement('div');
      div.className = 'io-item';
      
      // Беремо назву з глобального об'єкта за індексом
      const name = currentLabels[i];
      const displayLabel = name ? `#${i} <div class="pin-name">${name}</div>` : `#${i}`;
      
      div.innerHTML = `
        <small class="io-index">${displayLabel}</small>
                <div id="${dotClassPrefix}-${i}" class="io-dot"></div>
      `;
      grid.appendChild(div);
    });
  }

  dataArray.forEach((val, i) => {
    const dot = document.getElementById(`${dotClassPrefix}-${i}`);
    if (dot) {
      if (val === 1) dot.classList.add(colorClass);
      else dot.classList.remove(colorClass);
    }
  });
}

async function updateStatus() {
  try {
    const response = await fetch('/api/status');
    const data = await response.json();

    // 1. Дані логіки
    document.getElementById('val-mode').innerText = "MODE " + data.mode;
    document.getElementById('val-active-op').innerText = data.active_operation || "IDLE";
    document.getElementById('val-encoder').innerText = data.encoder_value + '°';
    document.getElementById('val-cycle').innerText = data.read_cycle_ms + ' ms';

    // 2. Логіка Safety Lock
    const isLocked = data.is_safety_locked;
    const sLabel = document.getElementById('val-safety');
    const sCard = document.getElementById('safety-card');
    const outContainer = document.getElementById('outputs-container');
    const hOut = document.getElementById('header-out');
    const sOut = document.getElementById('sync-outputs');

    if (isLocked) {
      sLabel.innerText = "ЗАБЛОКОВАНО";
      sLabel.className = "fw-bold text-danger";
      sCard.classList.add("border-danger");

      // Стилі для виходів (LOCKED)
      hOut.className = "card-header d-flex justify-content-between align-items-center header-locked py-2";
      sOut.innerText = "HARDWARE LOCK";
      sOut.className = "badge bg-dark text-white";
      outContainer.classList.add("outputs-locked");
      document.getElementById('val-stop-reason').innerText = data.stop_reason;
    } else {
      sLabel.innerText = "ГОТОВНІСТЬ";
      sLabel.className = "fw-bold text-success";
      sCard.classList.remove("border-danger");

      // Стилі для виходів (READY / ONLINE)
      hOut.className = "card-header d-flex justify-content-between align-items-center header-ready py-2";
      updateStatusLabel('sync-outputs', !data.is_safety_locked);
      outContainer.classList.remove("outputs-locked");
      document.getElementById('val-stop-reason').innerText = "";
    }

    // 3. Статуси модулів
    updateStatusLabel('sync-inputs', data.is_inputs_online);
    const encLabel = document.getElementById('sync-encoder');
    encLabel.innerText = data.is_encoder_online ? "ENC OK" : "ENC ERROR";
    encLabel.className = data.is_encoder_online ? "small fw-bold mt-1 text-success" : "small fw-bold mt-1 text-danger";

    // 4. Сітки
    renderGrid('inputs-grid', data.device10_in, 'dot-in', 'dot-in-on');
    renderGrid('outputs-grid', data.device20_out, 'dot-out', 'dot-out-on');

    document.getElementById('connection-badge').className = "badge bg-success";
    document.getElementById('connection-badge').innerText = "Зв'язок OK";

  } catch (err) {
    document.getElementById('connection-badge').className = "badge bg-danger";
    document.getElementById('connection-badge').innerText = "Зв'язок ВТРАЧЕНО";
  }
}

init();
