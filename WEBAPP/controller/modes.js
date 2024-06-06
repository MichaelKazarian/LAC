import { Communicator } from "./communicator.js";
import { manual, operation } from "./operations.js";

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
  _intervalVal = 100
  _inputState
  _onStopped
  _operation
  _cycleState;
#task
  constructor() {
    this._description = "Base mode";
    this._id = MODE;
    this._communicator = new Communicator();
    this._outputState = {
      r7: [1, 1, 1],
      r8: [0, 0, 0],
      r9: [1, 1, 0],
      r10: [1, 0, 1],
      r11: [0, 1, 1]
    };
    this.init();
  }

  init() {
    this._inputState = {
      type: "input",
      degree: 0,
      rawinput: "[]",
      quantity: 0,
      manualOperations: [],
      error: ""
    };
  }

  get description() {
    return this._description;
  }

  get id() {
    return this._id;
  }

  /**
   * After stop callback
   */
  set onStopped(f) {
    this._onStopped = f;
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
      if (data) {
        this._inputState.rawinput = (data.data);
        this._inputState.degree++;
        if (this._inputState.degree > 719) this._inputState.degree = 0;
        process.send(JSON.stringify(this._inputState));
      }
      else {
        console.log("READ error", data);
        await sleep(50);
      } // more than time of readHoldingRegisters to avoid crashes
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

  //Send stop message to this.#device
  _sendStop = async () => {
    await this._write([1, 0, 1]);
  };

  async operate () {
    // console.log("IS", this._inputState);
    this.nextPiece();
    if (this.isPieceWritable()) {
      let r = this.parseOperationMessage(this._operation.value.operation);
      this.addTask(async () => {
        // console.log("WRITE", r);
        await this._write(r);
      });
    } else {
      await this.addTask(this._read);
    }
    if (this.isOperationDone()) {
      this.startNextOperation();
    }
  }

  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    this._mainInterval = setInterval(() => {
      this.operate();
    }, this._intervalVal);
    return true;
  }

  addTask(task) {
    this._communicator.addTask(task);
  }

  /**
   * Public stop message. Calls mode-based stop implementations.
   */
  async stop() {
    await this._stop();
  }

  /**
   * Private stop executor.
   */
  async _stop() {
    await clearInterval(this._mainInterval);
    await this._communicator.addTask(this._sendStop);
    await console.log(`${this._description} stop`);
  }

  /**
   * Parse message to recognize current operation.
   * @param {String} msg message like radio&r10. r10 is a instuction.
   */
  parseOperationMessage = (msg) => {
    let id = msg.split("&")[1];
    return this._outputState[id];
  }

  /**
   * Runs next operation if exists. Stop cycle otherwise.
   */
  async startNextOperation() {
    if (this._operation.value !== undefined) {      // Start next operation
      this._cycleState = this._operation.value();
    } else {                                        // Cycle done
      await this._stop();
    }
  }

  /**
   * Set next piece of the operation.
   */
  nextPiece(){
    this._operation = this._cycleState.next(this._inputState);
  }

  /**
   * Returns state of current operation.
   * @return true if current operation done; false otherwise.
   */
  isOperationDone() {
    return this._operation.done;
  }

  /**
   * Returns state of current operation.
   * @return true if current operation done; false otherwise.
   */
  isPieceWritable() {
    return !this.isOperationDone() && this._operation.value.type === "write";
  }
}

//*****************************************************

class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
    this._id = MODE_MANUAL;
  }

  init() {
    super.init();
    this._cycleState = manual();
  }

  nextPiece(){
    super.nextPiece();
    this._inputState.manualOperations = this._operation.value["available"];
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

//*****************************************************

class ModeСycle extends Mode {
  _manualStop
  constructor() {
    super();
    this._description = "Cycle mode";
    this._id = MODE_CYCLE;
    this.init();
  }

  init() {
    super.init();
    this._cycleState = operation();
    this._manualStop = false;
  }
}

//*****************************************************

class ModeOnceСycle extends ModeСycle {
  constructor() {
    super();
    this._description = "Once cycle mode";
    this._id = MODE_ONCE_CYCLE;
  }

  /**
   * Immediately terminate the cycle and current operation.
   * Executes onStopped() behavior if present.
   */
  async stop() {
    // console.log("STOP MANUALLY");
    this._manualStop = true;
    await this._stop();
    }

  async _stop() {
    await super._stop();
    if (this._operation.value === undefined || this._manualStop) {
      if (this._onStopped !== undefined) await this._onStopped();
    }
  }
}

//*****************************************************

class ModeAuto extends ModeСycle {
  _beforeStop
  _quantity
  constructor() {
    super();
    this._description = "Automatic mode";
    this._id = MODE_AUTO;
    this._quantity = 0;
  }

  /**
   * Before stop callback
   */
  set beforeStop(f) {
    this._beforeStop = f;
  }

  async operate() {
    await super.operate();
    this._inputState.quantity = this._quantity;
    if (!this._manualStop
        && this._operation.done
        && this._operation.value === undefined) { // cycle done
      this.init();
      this._quantity++;
      await sleep(2000);
      await this.activate();
    }
  }

  /**
   * Runs next operation if exists. Stop the cycle otherwise.
   * Also, terminates the cycle after the current operation is done.
   */
  async startNextOperation() {
    if (this._manualStop) {
      await this._stop();
    } else await super.startNextOperation();
  }

  /**
   * Executes beforeStop(), sends stop signal and waits for current operation
   * to be done to stop.
   */
  stop() {
    this._manualStop = true;
    if (this._beforeStop) this._beforeStop();
  }

  async _stop() {
    await super._stop();
    if (this._manualStop
        && this._onStopped !== undefined) {
      await this._onStopped();
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
