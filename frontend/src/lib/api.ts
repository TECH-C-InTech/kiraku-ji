const apiBaseUrl = () => (process.env.NEXT_PUBLIC_API_BASE ?? "").trim();

/**
 * API ベースURLを返し、未設定なら例外を投げる。
 */
export const getApiBaseUrl = () => {
  const trimmedUrl = apiBaseUrl();
  if (trimmedUrl === "") {
    throw new Error("NEXT_PUBLIC_API_BASE が未設定です");
  }
  return trimmedUrl;
};

/**
 * API ベースURL末尾のスラッシュを除去して返す。
 */
export const normalizeApiBaseUrl = () => getApiBaseUrl().replace(/\/+$/, "");
