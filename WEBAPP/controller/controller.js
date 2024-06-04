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
  _info
  constructor() {
    this.#allowInterval = true;
    this._setManual();
  }

  /**
   * Add the task to the current mode.
   */
  addTask(task) {
    this.#mode.addTask(task);
  }

  /**
   * Waits while the current mode stops and switches to another.
   * Note: Stop implementation is mode-based.
   */
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

  /** Switches to manual mode */
  _setManual() {
    this.#mode = new ModeManual();
    this._infoUpdate(this.#mode.activate()? "success": "error");
  }

  /** Switches to once-cycle mode */
  _setOnceCycle() {
    this.#mode = new ModeOnceСycle();
    this.#mode.onStopped = () => this._setManual();
    this._infoUpdate(this.#mode.activate()? "success": "error");
  }

  /** Switches to automatic mode */
  _setAuto() {
    this.#mode = new ModeAuto();
    this.#mode.onStopped = () => this._setManual();
    this._infoUpdate(this.#mode.activate()? "success": "error");
  }

  _infoUpdate(as) {
    this._info = {
      type: "mode",
      modeId: this.#mode.id,
      modeDescription: this.#mode.description,
      modeStatus: as
    };
    process.send(JSON.stringify(this._info));
  }

  /**
   * Sends stop command to current mode. Stop implementation is mode-based.
   */
  async stop() {
    await this.#mode.stop();
    // process.exit();
  }
}

export { Controller };
