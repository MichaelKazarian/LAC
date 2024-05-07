class Mode {
_description
  constructor() {
    this._description = "Base mode";
  }
  get description() {
    return this._description;
  }
}

class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
  }

}

class ModeOnceСycle extends Mode {
  constructor() {
    super();
    this._description = "Once cycle mode";
  }  
}

class ModeAuto extends ModeOnceСycle {
  constructor() {
    super();
    this._description = "Automatic mode";
  }
}

export {
  Mode,
  ModeManual,
  ModeOnceСycle,
  ModeAuto
};
