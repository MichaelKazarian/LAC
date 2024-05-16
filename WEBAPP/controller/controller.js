import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO,
  Mode, ModeManual, ModeOnceСycle, ModeAuto
} from "./modes.js";

/**
 * Class controls equipment mode.
 */
class Controller {
#allowInterval
#mainInterval
#mode
#intervalVal = 50
  constructor() {
    this.#allowInterval = true;
    this.#mode = new ModeManual();
  }

  addTask(task) {
    this.#mode.addTask(task);
  }

  run() {
    this.#mainInterval = setInterval(() => {
      this.addTask("read");
    }, this.#intervalVal);
  }

  setMode(mode) {
    this.#mode.stop();
    switch (mode) {
    case MODE_ONCE_CYCLE:
      this.#mode = new ModeOnceСycle();
      break;
    case MODE_AUTO:
      this.#mode = new ModeAuto();
      break;
    default:
      this.#mode = new ModeManual();
    }
    return new Promise((resolve, reject) => {
      let status = {
        type: "mode",
        modeId: this.#mode.id,
        modeDescription: this.#mode.description,
        modeStatus: "success"
      };
      if (this.#mode.activate() === true) {
        resolve(JSON.stringify(status));
      } else {
        status.modeStatus = "error";
        reject(JSON.stringify(status));
      }
    });
  }

  stop() {
    this.#mode.stop();
    return this.setMode(MODE_MANUAL);
    // clearInterval(this.#mainInterval);
    // process.exit();
  }
}

export { Controller };
