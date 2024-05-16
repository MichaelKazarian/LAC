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
    this.#outputState = {
      r9: 1,
      r10: 1,
      r11: 1
    };
  }

  get inputState() {
    return this.#inputState;
  }

  get device() {
    return this.#device;
  }

  /**
   * Write the values [0 ... 0xffff] to registers starting at address 3
   * It is similar to executing
   * mbpoll -0 -m rtu -a 20 -b 19200 -t 4 -r 3 -P none /dev/ttyUSB0 1 0 1
   * Note: Note: time of writeRegisters near 0.04 sec
   */
  write = async (data) => {
    await sleep(50); // more than time of writeRegisters to avoid crashes
    await this.#device.writeRegisters(3, data);
    // await sleep(50);
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
      if (this.#que.length > 1) console.log("que0:", this.#que);
      let task = this.#que.shift();
      task();
    }
  })
    .catch((e) => {
      console.error("ERRR: ", e.message);
    });
}
  
  /**
   * Parse message to recognize current operation.
   * @param {String} msg message like radio&r10. r10 is a instuction.
   */
  parseOperationMessage = (msg) => {
    let res = [];
    let ids = msg.split("&")[1];
    for (let i in this.#outputState) {
      if (i === ids) {
        this.#outputState[i] = Math.abs(this.#outputState[i]-1);
      }
      res.push(this.#outputState[i]);
    }
    return res;
  }
}

export { Communicator }
