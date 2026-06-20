package lscolor

// Small fallback for terminals without LS_COLORS. Prefer simple ANSI colors so
// user themes remain in control; richer palettes should come from LS_COLORS.
const defaultColors = `di=34:ln=36:ex=32:fi=0`
