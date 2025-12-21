"use client";

import { useRouter } from "next/navigation";
import ResultCard from "@/components/modal/result-card";

type FortuneResultProps = {
  defaultText?: string;
};

/** おみくじ結果ページの表示を組み立てる。 */
export default function FortuneResultPage({
  defaultText = "今日のきらくじ: ここに結果テキストが入ります。",
}: FortuneResultProps) {
  const router = useRouter();
  const resultText = defaultText;

  /** おみくじ結果を共有し、未対応ならコピーへ切り替える。 */
  const handleShare = async () => {
    const shareText = resultText;
    if (navigator.share) {
      try {
        await navigator.share({ text: shareText });
        return;
      } catch (error) {
        if (error instanceof DOMException && error.name === "AbortError") {
          return;
        }
      }
    }

    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(shareText);
        window.alert("共有用のテキストをコピーしました");
        return;
      } catch (_error) {
        window.alert("共有に失敗しました");
        return;
      }
    }

    window.prompt("共有用テキストをコピーしてください。", shareText);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 px-4 py-8 font-sans text-zinc-900 md:px-0">
      <main className="relative flex w-full max-w-lg flex-col gap-8 rounded-xl bg-zinc-900 px-6 py-10 text-center shadow-lg md:max-w-xl md:px-8 md:py-12">
        <ResultCard
          resultText={resultText}
          onRetry={() => router.push("/")}
          buttonLabel="Try again"
          buttonClassName="border-zinc-200 text-zinc-100"
          secondaryButtonLabel="Share"
          onSecondary={handleShare}
          secondaryButtonClassName="border-zinc-200 text-zinc-100"
          reverseButtons
        />
      </main>
    </div>
  );
}
