"use client";

import { useState } from "react";

type FortuneResultProps = {
  defaultText?: string;
};

export default function FortuneResultPage({
  defaultText = "今日のきらくじ: ここに結果テキストが入ります。",
}: FortuneResultProps) {
  const [resultText] = useState(defaultText);

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans text-zinc-900">
      <main className="flex w-full max-w-xl flex-col gap-6 rounded-3xl bg-white px-8 py-12 text-center shadow-lg">
        <p className="font-medium text-base">{resultText}</p>
        <button
          className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
          type="button"
        >
          もう一度懺悔する
        </button>
      </main>
    </div>
  );
}
