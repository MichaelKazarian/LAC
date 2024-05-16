import { Communicator } from "./communicator.js";

const MODE            = "mode";
const MODE_MANUAL     = "mode-manual";
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
  #task
  constructor() {
    this._description = "Base mode";
    this._id = MODE;
    this._communicator = new Communicator();
    this._inputState = {
      type: "input",
      degree: 51,
      rawinput: "[]",
      error: ""
    };
    this._outputState = {
      r9: 1,
      r10: 1,
      r11: 1
    };

  }
  get description() {
    return this._description;
  }

  get id() {
    return this._id;
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
    console.log("DO STOP");
    let sendStop = async () => {
      await this._write([1, 1, 1]);
    };
    this.addTask(sendStop);
  }

  // TODO merge with activate
  run() {
    this._mainInterval = setInterval(() => {
      this.addTask(this._read);
    }, this._intervalVal);
  }

  
  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    this.run();   // TODO merge with run
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

  /**
   * Parse message to recognize current operation.
   * @param {String} msg message like radio&r10. r10 is a instuction.
   */
  parseOperationMessage = (msg) => {
    let res = [];
    let ids = msg.split("&")[1];
    for (let i in this._outputState) {
      if (i === ids) {
        this._outputState[i] = Math.abs(this._outputState[i]-1);
      }
      res.push(this._outputState[i]);
    }
    return res;
  }

}

class ModeOnceСycle extends Mode {
  constructor() {
    super();
    this._description = "Once cycle mode";
    this._id = MODE_ONCE_CYCLE;
  }  
}

class ModeAuto extends ModeOnceСycle {
  constructor() {
    super();
    this._description = "Automatic mode";
    this._id = MODE_AUTO;
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
