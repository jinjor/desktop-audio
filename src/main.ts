import { app, BrowserWindow } from "electron";
import { AudioClient } from "./core/audio-client";
import path from "path";
import { ipcMain } from "electron";
import { createMenu } from "./core/menu";

createMenu();

(async () => {
  // set up AudioClient
  const audioClient = new AudioClient();
  audioClient.onConnected = () => {
    console.log("connected to the audio server");
  };
  audioClient.onDisconnected = () => {
    console.log("disconnected from the audio server");
    process.exit(1);
  };
  audioClient.onError = (e) => {
    console.log(e.message);
    process.exit(1);
  };
  audioClient.onMessage = (message) => {
    console.log("got message from go", message);
    win!.webContents.send("audio", message);
  };
  await audioClient.connect();

  // set up IPC
  ipcMain.on("audio", (_, command: string[]) => {
    console.log("got message from web", command);
    audioClient.send(command);
  });

  // set up App and Window
  let win: BrowserWindow | undefined;
  const appPath = path.resolve(__dirname, `ui/app.js`);
  const htmlPath = path.resolve(__dirname, `../index.html`);

  function createWindow() {
    win = new BrowserWindow({
      width: 800,
      height: 600,
      webPreferences: {
        nodeIntegration: false,
        contextIsolation: true,
        preload: appPath,
      },
    });
    win.loadFile(htmlPath);
  }
  app.whenReady().then(createWindow);
  app.on("window-all-closed", () => {
    if (process.platform !== "darwin") {
      app.quit();
    }
  });
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
