import FortuneResultPage from "@/components/pages/result";

type ResultPageProps = {
  searchParams?: {
    text?: string | string[];
  };
};

/**
 * おみくじ結果ページを描画する。
 */
export default function ResultPage({ searchParams }: ResultPageProps) {
  const textParam = searchParams?.text;
  const defaultText =
    typeof textParam === "string" && textParam.trim() !== ""
      ? textParam
      : undefined;

  return <FortuneResultPage defaultText={defaultText} />;
}
