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

  async setMode(mode) {
    await this.stop();
    switch (mode) {
    case MODE_ONCE_CYCLE:
      this._setOnceCycle();
      break;
    case MODE_AUTO:
      this._setAuto();
      break;
    default:
      this._setManual();
    }
  }

  _setManual() {
    this.#mode = new ModeManual();
    let status = {
      type: "mode",
      modeId: this.#mode.id,
      modeDescription: this.#mode.description,
      modeStatus: this.#mode.activate()? "success": "error"
    };
    process.send(JSON.stringify(status));
  }

  _setOnceCycle() {
    this.#mode = new ModeOnceСycle();
    this.#mode.onStopped = () => this._setManual();
    let status = {
      type: "mode",
      modeId: this.#mode.id,
      modeDescription: this.#mode.description,
      modeStatus: this.#mode.activate()? "success": "error"
    };
    process.send(JSON.stringify(status));
  }

  _setAuto() {
    this.#mode = new ModeAuto();
    this.#mode.onStopped = () => this._setManual();
    let status = {
      type: "mode",
      modeId: this.#mode.id,
      modeDescription: this.#mode.description,
      modeStatus: this.#mode.activate()? "success": "error"
    };
    process.send(JSON.stringify(status));
  }

  async stop() {
    await this.#mode.stop();
    // process.exit();
  }
}

export { Controller };
