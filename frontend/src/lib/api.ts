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

export type ApiErrorResponse = {
  message?: string;
};

const statusMessageMap: Record<number, string> = {
  400: "入力が正しくありません",
  404: "おみくじがまだ準備できていません",
  409: "同じ内容が登録済みです",
  500: "サーバーで問題が発生しました。しばらく待って再試行してください",
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
