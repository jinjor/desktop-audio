import React from "react";
import { Action } from "./state";
import * as d from "./decoder";
import { Select } from "./select";
import { ScheduleFn } from "./react-util";
import { Dialog } from "./modal";

export const presetMetaDecoder = d.object({
  name: d.string(),
});
export type PresetMeta = d.TypeOf<typeof presetMetaDecoder>;
export type PresetState = {
  list: PresetMeta[];
  name: string;
  saver: { name: string } | null;
  remover: { name: string } | null;
};
export type PresetAction =
  | { type: "selectPreset"; value: string }
  | { type: "openSaver"; value: string }
  | { type: "inputName"; value: string }
  | { type: "save" }
  | { type: "closeSaver" }
  | { type: "openRemover"; value: string }
  | { type: "remove" }
  | { type: "closeRemover" };
export const presetReducer = (
  state: PresetState,
  action: PresetAction,
  effect: ScheduleFn<Action>
): PresetState => {
  switch (action.type) {
    case "selectPreset":
      effect((schedule) =>
        schedule({ type: "loadPreset", value: action.value })
      );
      return { ...state, name: action.value };
    case "inputName":
      return { ...state, saver: { ...state.saver, name: action.value } };
    case "openSaver":
      return {
        ...state,
        saver: { ...state.saver, name: action.value },
      };
    case "save":
      effect((schedule) =>
        schedule({ type: "savePreset", value: state.saver!.name })
      );
      return {
        ...state,
        name: state.saver!.name,
        saver: null,
      };
    case "closeSaver":
      return { ...state, saver: null };
    case "openRemover":
      return {
        ...state,
        remover: { ...state.remover, name: action.value },
      };
    case "remove":
      effect((schedule) =>
        schedule({ type: "removePreset", value: state.remover!.name })
      );
      return {
        ...state,
        name: "",
        remover: null,
      };
    case "closeRemover":
      return { ...state, remover: null };
  }
};
export const initialPresetState: PresetState = {
  list: [],
  name: "",
  saver: null,
  remover: null,
};
export const setPresetList = (
  state: PresetState,
  list: PresetMeta[]
): PresetState => {
  return { ...state, list };
};
export const Presets = React.memo(
  (o: { state: PresetState; dispatch: React.Dispatch<PresetAction> }) => {
    const onChangePreset = (value: string) =>
      o.dispatch({ type: "selectPreset", value });
    return (
      <div>
        <PresetSelect
          list={o.state.list}
          onChange={onChangePreset}
          value={o.state.name ?? ""}
        />
        <button
          onClick={() => o.dispatch({ type: "openSaver", value: o.state.name })}
        >
          Save
        </button>
        <button
          disabled={!o.state.name}
          onClick={() =>
            o.dispatch({ type: "openRemover", value: o.state.name })
          }
        >
          Remove
        </button>
        <PresetSaver state={o.state.saver} dispatch={o.dispatch} />
        <PresetRemover state={o.state.remover} dispatch={o.dispatch} />
      </div>
    );
  }
);
const PresetSelect = React.memo(
  (o: {
    list: PresetMeta[];
    onChange: (value: string) => void;
    value: string;
  }) => {
    // TODO: avoid selecting the first item
    const list = ["", ...o.list.map((item) => item.name)];
    return <Select list={list} value={o.value} onChange={o.onChange} />;
  }
);
const PresetSaver = React.memo(
  (o: {
    state: { name: string } | null;
    dispatch: React.Dispatch<PresetAction>;
  }) => {
    const onChangeName = (e: React.ChangeEvent<HTMLInputElement>) => {
      o.dispatch({ type: "inputName", value: e.target.value });
    };
    const onClose = () => {
      o.dispatch({ type: "closeSaver" });
    };
    const onOK = () => {
      o.dispatch({ type: "save" });
    };
    const onCancel = () => {
      o.dispatch({ type: "closeSaver" });
    };
    if (o.state == null) {
      return null;
    }
    return (
      <Dialog
        title="Save"
        onClose={onClose}
        onCancel={onCancel}
        onOK={onOK}
        okDisabled={o.state.name.trim() === ""}
      >
        <label>
          Name:
          <input onChange={onChangeName} value={o.state.name} />
        </label>
      </Dialog>
    );
  }
);

const PresetRemover = React.memo(
  (o: {
    state: { name: string } | null;
    dispatch: React.Dispatch<PresetAction>;
  }) => {
    const onClose = () => {
      o.dispatch({ type: "closeRemover" });
    };
    const onOK = () => {
      o.dispatch({ type: "remove" });
    };
    const onCancel = () => {
      o.dispatch({ type: "closeRemover" });
    };
    if (o.state == null) {
      return null;
    }
    return (
      <Dialog title="Remove" onClose={onClose} onCancel={onCancel} onOK={onOK}>
        <label>Remove {o.state.name}?</label>
      </Dialog>
    );
  }
);
