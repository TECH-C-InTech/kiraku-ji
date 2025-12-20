"use client";

import { useRouter } from "next/navigation";
import ResultCard from "@/components/modal/result-card";

type FortuneResultProps = {
  defaultText?: string;
};

export default function FortuneResultPage({
  defaultText = "今日のきらくじ: ここに結果テキストが入ります。",
}: FortuneResultProps) {
  const router = useRouter();
  const resultText = defaultText;

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans text-zinc-900">
      <main className="flex w-full max-w-xl flex-col gap-6 rounded-3xl bg-white px-8 py-12 text-center shadow-lg">
        <ResultCard
          resultText={resultText}
          onRetry={() => router.push("/")}
          buttonLabel="もう一度懺悔する"
        />
      </main>
    </div>
  );
}
