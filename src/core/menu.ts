import { app, shell } from "electron";
import { Menu } from "electron";

const template: Electron.MenuItemConstructorOptions[] = [
  ...(process.platform === "darwin"
    ? [
        {
          label: app.name,
          submenu: [
            {
              label: "About " + app.name,
              role: "about" as const,
            },
            {
              type: "separator" as const,
            },
            {
              label: "Services",
              role: "services" as const,
              submenu: [],
            },
            {
              type: "separator" as const,
            },
            {
              label: "Hide " + app.name,
              accelerator: "Command+H",
              role: "hide" as const,
            },
            {
              label: "Hide Others",
              accelerator: "Command+Shift+H",
              role: "hideOthers" as const,
            },
            {
              label: "Show All",
              role: "unhide" as const,
            },
            {
              type: "separator" as const,
            },
            {
              label: "Quit",
              accelerator: "Command+Q",
              click: function () {
                app.quit();
              },
            },
          ],
        },
      ]
    : []),
  {
    label: "View",
    submenu: [
      {
        label: "Toggle Developer Tools",
        accelerator: (function () {
          if (process.platform === "darwin") return "Alt+Command+I";
          else return "Ctrl+Shift+I";
        })(),
        click: function (item: any, focusedWindow: any) {
          if (focusedWindow) focusedWindow.toggleDevTools();
        },
      },
    ],
  },
  {
    label: "Window",
    role: "window" as const,
    submenu: [
      {
        label: "Minimize",
        accelerator: "CmdOrCtrl+M",
        role: "minimize" as const,
      },
      {
        label: "Close",
        accelerator: "CmdOrCtrl+W",
        role: "close" as const,
      },
      ...(process.platform === "darwin"
        ? [
            {
              type: "separator" as const,
            },
            {
              label: "Bring All to Front",
              role: "front" as const,
            },
          ]
        : []),
    ],
  },
  {
    label: "Help",
    role: "help",
    submenu: [
      {
        label: "Learn More",
        click: function () {
          shell.openExternal("http://electron.atom.io");
        },
      },
    ],
  },
];

export function createMenu(): void {
  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}
