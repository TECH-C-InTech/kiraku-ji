const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE ?? "";

/**
 * API ベースURLを返し、未設定なら例外を投げる。
 */
export const API_BASE_URL = (() => {
  const trimmedUrl = apiBaseUrl.trim();
  if (trimmedUrl === "") {
    throw new Error("NEXT_PUBLIC_API_BASE が未設定です");
  }
  return trimmedUrl;
})();
