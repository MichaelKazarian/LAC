import { Controller } from "./controller.js";
import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO
} from "./modes.js";

let controller = new Controller();
controller.run();

process.on('message', (msg) => {
  console.log('Message from parent:', msg, typeof(msg));
  if (msg === "STOP") {
    controller.stop();
  }
  if (msg === "read") {
    controller.addTask("read");
  }
  if (msg.startsWith("radio&")) {
    controller.addTask(msg);
  }
  if (msg.startsWith("mode-")) {
    controller.setMode(msg)
      .then(status => process.send(status))
      .catch(status => process.send(status));
  }
});

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
