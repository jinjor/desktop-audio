import React from "react";

export const Radio = (o: {
  list: string[];
  value: string;
  onChange: (value: string) => void;
}) => {
  return (
    <div style={{ display: "flex", flexFlow: "column", userSelect: "none" }}>
      {o.list.map((item) => {
        const selected = item === o.value;
        return (
          <Option
            key={item}
            label={item}
            selected={selected}
            onClick={() => {
              if (!selected) {
                o.onChange(item);
              }
            }}
          />
        );
      })}
    </div>
  );
};

const Option = (o: {
  label: string;
  selected: boolean;
  onClick: () => void;
}) => {
  return (
    <div
      style={{ display: "inline-flex", alignItems: "center" }}
      onClick={() => o.onClick()}
    >
      <div
        style={{
          width: "7px",
          height: "7px",
          borderRadius: "50%",
          border: "solid 2px #222",
          backgroundColor: o.selected ? "rgb(153,119,255)" : "#000",
          marginRight: "2px",
        }}
      />
      <div>{o.label}</div>
    </div>
  );
};
