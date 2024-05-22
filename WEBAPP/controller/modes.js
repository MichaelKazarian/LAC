import { Communicator } from "./communicator.js";
import { operation } from "./operations.js";

const MODE            = "mode";
const MODE_MANUAL     = "mode-manual";
const MODE_CYCLE      = "mode-cycle";
const MODE_ONCE_CYCLE = "mode-once-cycle";
const MODE_AUTO       = "mode-auto";


function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}


class Mode {
  _id;
  _description;
  _communicator
  _mainInterval
  _intervalVal = 50
  _inputState
  _cycleState
  _onX
  _operation
  #task
  constructor() {
    this._description = "Base mode";
    this._id = MODE;
    this._communicator = new Communicator();
    this._cycleState = operation();
    this._inputState = {
      type: "input",
      degree: 0,
      rawinput: "[]",
      error: ""
    };
    this._outputState = {
      r7: [1, 1, 1],
      r8: [0, 0, 0],
      r9: [1, 1, 0],
      r10: [1, 0, 1],
      r11: [0, 1, 1]
    };

  }

  get description() {
    return this._description;
  }

  get id() {
    return this._id;
  }

  set onX(f) {
    this._onX = f;
  }

  /**
   * Read the 3 registers starting at address 0 on device number this.#device
   * It is similar to executing
   * mbpoll -0 -m rtu -a 20 -b 9600 -t 4 -r 0 -c 3 -P none /dev/ttyUSB0
   * Note: time of readHoldingRegisters near 0.03 sec
   */
  _read = async () => {
    let d = this._communicator.device;
    await d.readHoldingRegisters(0, 3, async (err, data) => {
      if (data != null) {
        this._inputState.rawinput = (data.data);
        this._inputState.degree++;
        if (this._inputState.degree > 719) this._inputState.degree = 0;
        process.send(JSON.stringify(this._inputState));
      }
      else await sleep(50); // more than time of readHoldingRegisters to avoid crashes
    });
  }

  /**
   * Write the values [0 ... 0xffff] to registers starting at address 3
   * It is similar to executing
   * mbpoll -0 -m rtu -a 20 -b 19200 -t 4 -r 3 -P none /dev/ttyUSB0 1 0 1
   * Note: Note: time of writeRegisters near 0.04 sec
   */
  _write = async (data) => {
    await sleep(50); // more than time of writeRegisters to avoid crashes
    await this._communicator.device.writeRegisters(3, data);
    // await sleep(50);
  }

  _stop = () => {
    // console.log("DO STOP");
    let sendStop = async () => {
      await this._write([1, 0, 1]);
    };
    this.addTask(sendStop);
  }

  operate() {}

  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    this._mainInterval = setInterval(() => {
      this.addTask(this._read);
      this.operate();
    }, this._intervalVal);
    return true;
  }

  addTask(task) {
    this._communicator.addTask(task);
  }
  
  stop() {
    this._communicator.addTask(this._stop);
    clearInterval(this._mainInterval);
    console.log(`${this._description} stop`);
  }

  /**
   * Parse message to recognize current operation.
   * @param {String} msg message like radio&r10. r10 is a instuction.
   */
  parseOperationMessage = (msg) => {
    let id = msg.split("&")[1];
    return this._outputState[id];
  }
}

//*****************************************************
class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
    this._id = MODE_MANUAL;
  }

  addTask(task) {
    if (typeof(task) === "string" && task.startsWith("radio&")) {
      let r = this.parseOperationMessage(task);
      super.addTask(async () => {
        await this._write(r);
      });
      return;
    };
    super.addTask(task);
  }
}

class ModeСycle extends Mode {
  constructor() {
    super();
    this._description = "Cycle mode";
    this._id = MODE_CYCLE;
  }

  operate () {
    this._operation = this._cycleState.next(this._inputState);
    if (!this._operation.done) {
      if (this._operation.value.type === "write") {
        let r = this.parseOperationMessage(this._operation.value.operation);
        super.addTask(async () => {
          console.log("r", r);
          await this._write(r);
        });
      }
    } else {
      if (this._operation.value !== undefined) { // start next operation
        this._cycleState = this._operation.value();
      } else { // cycle done
        this.stop();
      }
    }
  }
}

class ModeOnceСycle extends ModeСycle {
  constructor() {
    super();
    this._description = "Once cycle mode";
    this._id = MODE_ONCE_CYCLE;
  }
 
  stop () {
    super.stop();
    if (this._operation.value === undefined) {
      if (this._onX !== undefined) this._onX();
    }
  }
}

class ModeAuto extends ModeСycle {
  constructor() {
    super();
    this._description = "Automatic mode";
    this._id = MODE_AUTO;
  }

  async operate() {
    await super.operate();
    if (this._operation.done && this._operation.value === undefined) { // cycle done
      this._inputState = {
        type: "input",
        degree: 0,
        rawinput: "[]",
        error: ""
      };
      this._cycleState = operation();
      await sleep(2000);
      await this.activate();
    }
  }
}

export {
  Mode,
  ModeManual,
  ModeOnceСycle,
  ModeAuto,
  MODE_MANUAL,
  MODE_ONCE_CYCLE,
  MODE_AUTO
};
