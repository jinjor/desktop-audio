import React from "react";

const waveNameToIconParams = (value: string) => {
  switch (value) {
    case "sine":
      return sine;
    case "triangle":
      return triangle;
    case "square":
    case "square-wt":
      return square;
    case "pulse":
      return pulse;
    case "saw":
    case "saw-wt":
      return saw;
    case "saw-rev":
      return sawRev;
    case "noise":
      return noise;
    default:
      throw new Error("unknown wave name: " + value);
  }
};

const sine = {
  d: "M -8,0 Q -4 -13 0 0 T 8 0",
  "stroke-width": 2.2,
};
const triangle = {
  d: "M -8,0 L -4,-7 4,7 8,0",
  "stroke-width": 2,
};
const square = {
  d: "M -8,0 L -8,-6 0,-6 0,6 8,6 8,0",
  "stroke-width": 2,
};
const pulse = {
  d: "M -8,0 L -8,-6 -4,-6 -4,6 8,6 8,0",
  "stroke-width": 2,
};
const saw = {
  d: "M -8,6 L 7,-6 7,6",
  "stroke-width": 2,
};
const sawRev = {
  d: "M -7,6 L -7,-6 8,6",
  "stroke-width": 2,
};
const noise = {
  d: "M -8,0 L -6,-2 -4,6 -2,-8 0,5 2,-5 4,7 6,-4 8,2",
  "stroke-width": 2,
};

export const WaveSelect = (o: {
  list: string[];
  value: string;
  columns?: number;
  onChange: (value: string) => void;
}) => {
  const columns = o.columns || 1;
  const size = 16;
  return (
    <div
      style={{
        display: "inline-grid",
        gridTemplateColumns: `repeat(${columns},1fr)`,
        gap: "4px 6px",
      }}
    >
      {o.list.map((item) => {
        const selected = item === o.value;
        const params = waveNameToIconParams(item);
        return (
          <Option
            size={size}
            selected={selected}
            params={params}
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
  size: number;
  selected: boolean;
  params: any;
  onClick: () => void;
}) => {
  const stroke = o.selected ? "rgb(153,119,255)" : "#111";
  return (
    <svg
      style={{ verticalAlign: "middle" }}
      width={o.size}
      height={o.size}
      viewBox="-10 -10 20 20"
      onClick={o.onClick}
    >
      <path {...o.params} fill="none" stroke={stroke} />;
    </svg>
  );
};
