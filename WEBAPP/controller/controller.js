import { Communicator } from "./communicator.js";

/**
 * Class controls equipment mode.
 */
class Controller {
#communicator
#allowInterval
#mainInterval
#intervalVal = 50
  constructor() {
    this.#communicator = new Communicator();
    this.#allowInterval = true;
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
    
  }

  stop() {
    clearInterval(this.#mainInterval);
    process.exit();
  }
}

export { Controller };
