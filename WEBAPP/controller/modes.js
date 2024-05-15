const MODE            = "mode";
const MODE_MANUAL     = "mode-manual";
const MODE_ONCE_CYCLE = "mode-once-cycle";
const MODE_AUTO       = "mode-auto";

class Mode {
  _id;
  _description;
  constructor() {
    this._description = "Base mode";
    this._id = MODE;
  }
  get description() {
    return this._description;
  }

  get id() {
    return this._id;
  }

  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    return true;
  }

  stop() {
    console.log("Mode stop");
  }
}

class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
    this._id = MODE_MANUAL;
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
