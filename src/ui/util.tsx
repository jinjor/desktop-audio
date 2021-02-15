import { useState } from "react";

export const useHover = () => {
  const [hovered, setHovered] = useState(false);
  const listeners = {
    onMouseEnter: () => setHovered(true),
    onMouseLeave: () => setHovered(false),
  };
  return [hovered, listeners];
};
