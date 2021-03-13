import React from "react";
import { colors } from "./styles";
import { useHover } from "./util";

export const Modal = (o: { onClose: () => void; children: any }) => {
  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        display: "flex",
        zIndex: 1,
      }}
    >
      <div
        style={{
          backgroundColor: "rgba(0,0,0,0.5)",
          position: "absolute",
          top: 0,
          left: 0,
          bottom: 0,
          right: 0,
        }}
        onClick={o.onClose}
      ></div>
      <div
        style={{
          backgroundColor: colors.background,
          margin: "auto",
          zIndex: 1,
        }}
      >
        {o.children}
      </div>
    </div>
  );
};

export const Dialog = React.memo(
  (o: {
    title: string;
    children: any;
    onClose: () => void;
    onCancel?: () => void;
    onOK?: () => void;
    okDisabled?: boolean;
  }) => {
    return (
      <Modal onClose={o.onClose}>
        <div
          style={{
            padding: "10px",
            display: "flex",
            justifyContent: "space-between",
          }}
        >
          <span>{o.title}</span>
          <CloseButton onClick={o.onClose} />
        </div>
        <div style={{ padding: "10px" }}>{o.children}</div>
        <div style={{ padding: "10px", textAlign: "right" }}>
          {o.onCancel != null ? (
            <button onClick={o.onCancel}>Cancel</button>
          ) : null}
          {o.onOK != null ? (
            <button onClick={o.onOK} disabled={o.okDisabled}>
              OK
            </button>
          ) : null}
        </div>
      </Modal>
    );
  }
);

const CloseButton = (o: { onClick: () => void }) => {
  const [hovered, listeners] = useHover();
  return (
    <svg
      {...listeners}
      onClick={o.onClick}
      style={{
        width: `14px`,
        height: `14px`,
        cursor: "pointer",
      }}
      viewBox={`-10 -10 20 20`}
      stroke={hovered ? "#aaa" : colors.textBase}
      strokeWidth={2}
    >
      <path d="M-8,-8 L8,8" />
      <path d="M-8,8 L8,-8" />
    </svg>
  );
};
