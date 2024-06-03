import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO,
  Mode, ModeManual, ModeOnceСycle, ModeAuto
} from "./modes.js";

/**
 * Class controls equipment mode.
 */
class Controller {
#allowInterval
#mode
  constructor() {
    this.#allowInterval = true;
    this.#mode = new ModeManual();
  }

  /**
   * Add the task to the current mode.
   */
  addTask(task) {
    this.#mode.addTask(task);
  }

  setMode(mode, manualStop=true) {
    if (manualStop) this.#mode.stop();
    switch (mode) {
    case MODE_ONCE_CYCLE:
      this.#mode = new ModeOnceСycle();
      this.#mode.onStopped = () => {
        this.setMode(MODE_MANUAL, false)
          .then(status => process.send(status))
          .catch(status => process.send(status));
      };
      break;
    case MODE_AUTO:
      this.#mode = new ModeAuto();
      this.#mode.onStopped = () => {
        this.setMode(MODE_MANUAL, false)
          .then(status => process.send(status))
          .catch(status => process.send(status));
      };
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
    //this.#mode.stop();
    if (this.#mode.id === "mode-auto") {
      this.#mode.sendStop();
      return new Promise((resolve, reject) => {
        let status = {
          type: "mode",
          modeId: this.#mode.id,
          modeDescription: this.#mode.description,
          modeStatus: "sended-stop"
        };
        resolve(JSON.stringify(status));
    });
    } else return this.setMode(MODE_MANUAL, true);
    // process.exit();
  }
}

export { Controller };
