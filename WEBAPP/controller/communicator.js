import  { Mode, ModeManual, ModeOnceСycle, ModeAuto }  from "./modes.js";
import  ModbusRTU from "modbus-serial";
import AsyncLock from 'async-lock';

var lock = new AsyncLock();

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

/**
 * Input/output Boards communicator via Modbus RTU.
 */
class Communicator {
#inputState;
#outputState;
#device
#que
  /**
   * @constructs
   * @param {Number} modbusId - slave Id. Default is 20.
   */
  constructor(modbusId=20) {
    if (Communicator._instance) {
      return Communicator._instance;
    }
    this.connectionInit(modbusId);
    this.#que = [];
    Communicator._instance = this;
  }

  /**
   * Open connection to a serial port
   */
  connectionInit(modbusId) {
    this.#device = new ModbusRTU();
    this.#device.connectRTUBuffered("/dev/ttyUSB0", { baudRate: 9600, debug: true });
    this.#device.setTimeout(500);
    this.#device.setID(modbusId);
  }

  get inputState() {
    return this.#inputState;
  }

  get device() {
    return this.#device;
  }

  /**
   * Add and execute new task. First task executes first.
   * @param {String} task - "read" or "radio&rN"
   */
  addTask = (task) => {
    lock.acquire("key", () => {
      let lastIsSame = () => (this.#que.length > 0 && this.#que.at(-1) === task);
      if (!lastIsSame())
        this.#que.push(task);
    }).then(this.#do);
  }

  /**
   * Executes tasks from queue.
   */
#do = async () => {
  lock.acquire("key", async() => {
    while (this.#que.length > 0) {
      // if (this.#que.length > 1) console.log("que0:", this.#que);
      let task = this.#que.shift();
      await task();
    }
  })
    .catch((e) => {
      console.error("ERRR: ", e.message);
    });
}
}

export { Communicator }
