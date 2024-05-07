import { Communicator } from "./communicator.js";
import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO,
  Mode, ModeManual, ModeOnceСycle, ModeAuto
} from "./modes.js";

/**
 * Class controls equipment mode.
 */
class Controller {
#communicator
#allowInterval
#mainInterval
#mode
#intervalVal = 50
  constructor() {
    this.#communicator = new Communicator();
    this.#allowInterval = true;
    this.#mode = new ModeManual();
  }

  addTask(task) {
    this.#communicator.addTask(task);
  }

  run() {
    this.#mainInterval = setInterval(() => {
      this.addTask("read");
    }, this.#intervalVal);
  }

  setMode(mode) {
    switch (mode) {
    case MODE_ONCE_CYCLE:
      this.#mode = new ModeOnceCycle();
      break;
    case MODE_AUTO:
      this.#mode = new ModeAuto();
      break;
    default:
      this.#mode = new ModeManual();
    }
    return new Promise((resolve, reject) => {
      if (this.#mode.activate() === true) {
        resolve();
      } else reject();
    });
  }

  stop() {
    clearInterval(this.#mainInterval);
    process.exit();
  }
}

export { Controller };
