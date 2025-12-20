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

export type ApiErrorResponse = {
  message?: string;
};

const statusMessageMap: Record<number, string> = {
  400: "入力が正しくありません",
  404: "データが見つかりません",
  409: "すでに登録されています",
  500: "サーバーで問題が発生しました",
};

/**
 * API エラーメッセージを統一して返す。
 */
export const getApiErrorMessage = async (
  response: Response,
  fallbackMessage = "通信に失敗しました",
): Promise<string> => {
  const defaultMessage = statusMessageMap[response.status] ?? fallbackMessage;
  try {
    const data = (await response.json()) as ApiErrorResponse;
    if (data.message) {
      return data.message;
    }
  } catch {
    // 何も取得できない場合は既定文言で返す。
  }
  return defaultMessage;
};
