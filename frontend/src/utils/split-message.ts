/**
 * 表示文を指定の長さで2行に分割する。
 */
export const splitMessage = (message: string, maxChars: number) => {
  const normalized = message.trim();

  // 1行に収まる場合はそのまま返す
  if (normalized.length <= maxChars) {
    return [normalized];
  }

  const breakChars = ["、", "。", " ", "　"];

  // 2行に分ける位置を決める
  const breakIndex = breakChars.reduce((maxIndex, char) => {
    const index = normalized.lastIndexOf(char, maxChars);
    if (index === -1) {
      return maxIndex;
    }
    return Math.max(maxIndex, index + 1);
  }, -1);
  const resolvedBreakIndex =
    breakIndex <= 0 ? Math.min(maxChars, normalized.length) : breakIndex;

  const firstLine = normalized.slice(0, resolvedBreakIndex).trimEnd();
  const secondLine = normalized.slice(resolvedBreakIndex).trimStart();

  if (secondLine.length === 0) {
    return [firstLine];
  }

  return [firstLine, secondLine];
};
