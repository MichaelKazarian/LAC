// static/js/lac.js

import { 
  initBaseEvents, 
  applyBaseState, 
  setMessage 
} from './operator.js';

let prevOperationsState = [];
const pieSegment = document.getElementById("pie-segment");
const pieText = document.getElementById("pie-text");

// Оновлення кругової діаграми градусів
function updatePieChart(degrees) {
  if (!pieSegment || !pieText) return;
  const percentage = (degrees / 360) * 100;
  pieSegment.setAttribute("stroke-dasharray", `${percentage}, 100`);
  pieText.textContent = `${Math.round(degrees)}°`;
}

// Оновлення 18+ ручних кнопок
function updateManualButtons(operations) {
  if (!operations) return;
  operations.forEach((state, index) => {
    const opId = index + 1;
    if (state !== prevOperationsState[index]) {
      const btn = document.getElementById(`op-${opId}`);
      if (btn) {
        btn.checked = (state === 1);
      }
    }
  });
  prevOperationsState = [...operations];
}

// Прив'язка кліків для ручних кнопок (18+ шт)
function initManualButtons() {
  const btnContainer = document.getElementById("btnContainer");
  if (!btnContainer) return;

  btnContainer.addEventListener("change", (e) => {
    if (e.target && e.target.id.startsWith("op-")) {
      const opId = e.target.id.replace("op-", "");
      const isChecked = e.target.checked ? 1 : 0;
      
      fetch(`/radio?id=${opId}&val=${isChecked}`)
        .catch(() => setMessage("error", "Помилка відправки ручної команди"));
    }
  });
}

// Головний цикл наладчика
document.addEventListener("DOMContentLoaded", () => {
  initBaseEvents();
  initManualButtons();

  async function cabinetState() {
    try {
      const response = await fetch("/state");
      if (response.ok) {
        const json = await response.json();
        
        // 1. Викликаємо базове оновлення з operator.js
        applyBaseState(json);

        // 2. Додаємо наладку: Градуси та Ручні кнопки
        if (json["degrees"] !== undefined) {
          updatePieChart(json["degrees"]);
        }
        if (json["operations"]) {
          updateManualButtons(json["operations"]);
        }
      } else {
        setMessage("error", `Помилка сервера: ${response.status}`);
      }
    } catch (err) {
      setMessage("error", "Зв'язок з контролером відсутній!");
    }
    setTimeout(cabinetState, 100);
  }

  setTimeout(cabinetState, 100);
});
