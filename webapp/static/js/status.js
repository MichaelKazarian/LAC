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
  if (grid.children.length === 0) {
    dataArray.forEach((_, i) => {
      const div = document.createElement('div');
      div.className = 'io-item';
      div.innerHTML = `<small style="font-size:0.6rem; color:#888">#${i}</small><div id="${dotClassPrefix}-${i}" class="io-dot"></div>`;
      grid.appendChild(div);
    });
  }
  dataArray.forEach((val, i) => {
    const dot = document.getElementById(`${dotClassPrefix}-${i}`);
    if (val === 1) dot.classList.add(colorClass);
    else dot.classList.remove(colorClass);
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

setInterval(updateStatus, 250);
updateStatus();
