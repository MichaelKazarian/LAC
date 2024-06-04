import { Controller } from "./controller.js";
import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO
} from "./modes.js";

let controller = new Controller();

process.on('message', (msg) => {
  console.log('Message from parent:', msg, typeof(msg));
  if (msg === "stop") {
    controller.stop();
      // .then(status => process.send(status))
      // .catch(status => process.send(status));
  }
  if (msg.startsWith("radio&")) {
    controller.addTask(msg);
  }
  if (msg.startsWith("mode-")) {
    controller.setMode(msg);
      // .then(status => process.send(status))
      // .catch(status => process.send(status));
  }
});

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
