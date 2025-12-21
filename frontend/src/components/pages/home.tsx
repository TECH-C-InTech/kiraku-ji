"use client";

import Image from "next/image";
import { useRouter } from "next/navigation";
import {
  type ChangeEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import KirakujiTransitionOverlay, {
  KIRAKUJI_TRANSITION_MS,
} from "@/components/ui/kirakuji-transition-overlay";
import { fetchRandomDraw } from "@/lib/draws";
import { createPost } from "@/lib/posts";

type Step = "input" | "loading" | "ready" | "error";

/** 表示文を指定の長さで2行に分割する。 */
const splitMessage = (message: string, maxChars: number) => {
  const normalized = message.trim();
  if (normalized.length <= maxChars) {
    return [normalized];
  }

  const breakChars = ["、", "。", " ", "　"];
  let breakIndex = -1;

  for (const char of breakChars) {
    const index = normalized.lastIndexOf(char, maxChars);
    if (index === -1) {
      continue;
    }
    breakIndex = Math.max(breakIndex, index + 1);
  }

  if (breakIndex <= 0) {
    breakIndex = Math.min(maxChars, normalized.length);
  }

  const firstLine = normalized.slice(0, breakIndex).trimEnd();
  const secondLine = normalized.slice(breakIndex).trimStart();

  if (secondLine.length === 0) {
    return [firstLine];
  }

  return [firstLine, secondLine];
};

export default function HomePage() {
  const router = useRouter();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentStep, setCurrentStep] = useState<Step>("input");
  const [loadingOrigin, setLoadingOrigin] = useState<"input" | "ready" | null>(
    null,
  );
  const [content, setContent] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [isTransitioning, setIsTransitioning] = useState(false);
  const transitionStartedAtRef = useRef<number | null>(null);
  const transitionPromiseRef = useRef<Promise<void> | null>(null);
  const transitionResolveRef = useRef<(() => void) | null>(null);
  const triggerButtonRef = useRef<HTMLButtonElement | null>(null);
  const modalRef = useRef<HTMLElement | null>(null);
  const inputRef = useRef<HTMLTextAreaElement | null>(null);
  const defaultPostError = "投稿に失敗しました";
  const defaultDrawError = "おみくじの取得に失敗しました";
  const welcomeMessage =
    "ようこそ、きらくじへ。自分の闇を差し出すと、おみくじが引けます。";
  const [welcomeLine, welcomeLineNext] = splitMessage(welcomeMessage, 16);
  const inputMessage = "闇をここに書いてね。";
  const [inputLine, inputLineNext] = splitMessage(inputMessage, 16);
  const contentLength = content.length;
  const trimmedLength = content.trim().length;
  const isSubmitDisabled = trimmedLength === 0 || contentLength > 140;
  const isSubmitButtonDisabled = isSubmitDisabled || currentStep === "loading";
  const modalClassName =
    currentStep === "ready"
      ? "relative flex w-[90%] max-w-md flex-col gap-4 text-center md:max-w-lg"
      : "relative flex w-[90%] max-w-md flex-col gap-4 rounded-xl bg-[#d6adc8] p-4 shadow-lg md:max-w-lg md:p-6";

  /** 遷移アニメーションの表示を開始する。 */
  const startTransition = () => {
    transitionStartedAtRef.current = Date.now();
    transitionPromiseRef.current = new Promise((resolve) => {
      transitionResolveRef.current = resolve;
    });
    setIsTransitioning(true);
  };

  /** 遷移アニメーションを停止して状態を戻す。 */
  const stopTransition = () => {
    transitionStartedAtRef.current = null;
    transitionPromiseRef.current = null;
    transitionResolveRef.current = null;
    setIsTransitioning(false);
  };

  /** 遷移アニメーションの完了を通知する。 */
  const completeTransition = () => {
    transitionResolveRef.current?.();
    transitionResolveRef.current = null;
  };

  /** 最低表示時間を満たすまで待機する。 */
  const waitForTransition = async () => {
    const startedAt = transitionStartedAtRef.current ?? Date.now();
    const elapsed = Date.now() - startedAt;
    const remaining = Math.max(0, KIRAKUJI_TRANSITION_MS - elapsed);
    const timerPromise =
      remaining > 0
        ? new Promise((resolve) => {
            window.setTimeout(resolve, remaining);
          })
        : Promise.resolve();
    await Promise.race([
      transitionPromiseRef.current ?? Promise.resolve(),
      timerPromise,
    ]);
    transitionPromiseRef.current = null;
    transitionResolveRef.current = null;
  };

  const handleContentChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    setContent(event.target.value);
  };

  const handleRetry = useCallback((options?: { clearContent?: boolean }) => {
    if (options?.clearContent) {
      setContent("");
    }
    setLoadingOrigin(null);
    setErrorMessage("");
    setCurrentStep("input");
  }, []);

  const handleSubmit = async () => {
    if (isSubmitDisabled || currentStep === "loading") {
      return;
    }
    setLoadingOrigin("input");
    setCurrentStep("loading");
    setErrorMessage("");
    try {
      await createPost(content.trim());
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : defaultPostError,
      );
      setLoadingOrigin(null);
      setCurrentStep("error");
      return;
    }

    setLoadingOrigin(null);
    setCurrentStep("ready");
  };

  const handleDraw = async () => {
    if (currentStep !== "ready" || isTransitioning) {
      return;
    }
    setLoadingOrigin("ready");
    setCurrentStep("loading");
    setErrorMessage("");
    startTransition();
    try {
      const draw = await fetchRandomDraw();
      const query = new URLSearchParams({ text: draw.result });
      await waitForTransition();
      setIsModalOpen(false);
      handleRetry({ clearContent: true });
      router.push(`/result?${query.toString()}`);
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : defaultDrawError,
      );
      setLoadingOrigin(null);
      setCurrentStep("error");
      stopTransition();
    }
  };

  useEffect(() => {
    if (!isModalOpen) {
      return;
    }

    // モーダル表示時は入力欄へフォーカスを移動する
    inputRef.current?.focus();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        setIsModalOpen(false);
        handleRetry();
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const modal = modalRef.current;
      if (!modal) {
        return;
      }

      const focusableElements = modal.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
      );

      if (focusableElements.length === 0) {
        return;
      }

      const first = focusableElements[0];
      const last = focusableElements[focusableElements.length - 1];

      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
        return;
      }

      if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      triggerButtonRef.current?.focus();
    };
  }, [handleRetry, isModalOpen]);

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-zinc-50 px-0 py-0 font-sans text-zinc-900 md:px-0">
      <div className="pointer-events-none absolute inset-x-0 top-0 z-0 flex justify-center opacity-70">
        <div className="w-full md:max-w-xl">
          <Image
            src="/curtain.png"
            alt=""
            width={1440}
            height={480}
            className="h-auto w-full object-top"
            priority
          />
        </div>
      </div>
      <div className="pointer-events-none absolute inset-0 z-10">
        <Image
          src="/pome_illust_normal.png"
          alt="闇を投げるイラスト"
          fill
          sizes="100vw"
          className="object-cover"
          priority
        />
      </div>
      <div className="-translate-y-[10px] relative z-20 flex w-full max-w-xl flex-col items-center gap-4 px-4 md:px-0">
        {!isModalOpen && (
          <div className="-top-48 -translate-x-1/2 pointer-events-none absolute left-1/2 w-full max-w-xs border border-zinc-900/10 bg-white px-4 py-3 text-center text-sm text-zinc-900 shadow-sm">
            <span className="sr-only">{welcomeMessage}</span>
            <span aria-hidden="true" className="kirakuji-typing-line">
              {welcomeLine}
            </span>
            {welcomeLineNext && (
              <span
                aria-hidden="true"
                className="kirakuji-typing-line is-second"
              >
                {welcomeLineNext}
              </span>
            )}
          </div>
        )}
        <button
          className="rounded-3xl p-2 transition hover:scale-[1.02] focus-visible:outline focus-visible:outline-2 focus-visible:outline-zinc-900 focus-visible:outline-offset-4"
          type="button"
          ref={triggerButtonRef}
          aria-label="闇を投げる"
          title="闇を投げる"
          onClick={() => {
            setIsModalOpen(true);
            handleRetry({ clearContent: true });
          }}
        >
          {currentStep !== "ready" && (
            <>
              <Image
                src="/hurt_normal.png"
                alt=""
                width={220}
                height={220}
                className="kirakuji-idle h-auto w-[200px] md:w-[220px]"
                priority
              />
              <span className="sr-only">闇を投げる</span>
            </>
          )}
        </button>
      </div>

      {isModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center px-4 py-8"
          role="dialog"
          aria-modal="true"
          onPointerDown={() => {
            setIsModalOpen(false);
            handleRetry();
          }}
        >
          <main
            className={modalClassName}
            ref={modalRef}
            onPointerDown={(event) => event.stopPropagation()}
          >
            {(currentStep === "input" ||
              (currentStep === "loading" && loadingOrigin === "input")) && (
              <section className="relative flex flex-col gap-4 text-center">
                <div className="-top-48 -translate-x-1/2 absolute left-1/2 w-full max-w-xs border border-zinc-900/10 bg-white px-4 py-3 text-sm text-zinc-900 shadow-sm">
                  <span className="sr-only">{inputMessage}</span>
                  <span aria-hidden="true" className="kirakuji-typing-line">
                    {inputLine}
                  </span>
                  {inputLineNext && (
                    <span
                      aria-hidden="true"
                      className="kirakuji-typing-line is-second"
                    >
                      {inputLineNext}
                    </span>
                  )}
                </div>
                <textarea
                  className="min-h-30 w-full resize-none rounded-md bg-white px-4 py-3 text-sm outline-none disabled:cursor-not-allowed disabled:bg-zinc-50"
                  maxLength={140}
                  placeholder="今夜の闇をひとこと"
                  value={content}
                  onChange={handleContentChange}
                  ref={inputRef}
                  disabled={currentStep === "loading"}
                />
                <button
                  className="mx-auto rounded-md bg-white px-6 py-2 font-semibold disabled:cursor-not-allowed disabled:opacity-50"
                  type="button"
                  onClick={handleSubmit}
                  disabled={isSubmitButtonDisabled}
                >
                  <p className="text-lg text-zinc-900">send</p>
                </button>
              </section>
            )}

            {currentStep === "ready" && (
              <section className="relative flex flex-col gap-4 text-center">
                <div className="-top-48 -translate-x-1/2 absolute left-1/2 w-full max-w-xs border border-zinc-900/10 bg-white px-4 py-3 text-sm text-zinc-900 shadow-sm">
                  <span className="sr-only">
                    きらくじを受け取る準備ができました。きらくじを受け取りますか？
                  </span>
                  <span aria-hidden="true" className="kirakuji-typing-line">
                    きらくじを受け取る準備ができました。
                  </span>
                  <span
                    aria-hidden="true"
                    className="kirakuji-typing-line is-second"
                  >
                    きらくじを受け取りますか？
                  </span>
                </div>
                <button
                  className="mx-auto rounded-3xl p-2 transition hover:scale-[1.02] focus-visible:outline focus-visible:outline-2 focus-visible:outline-zinc-900 focus-visible:outline-offset-4"
                  type="button"
                  aria-label="闇を引く"
                  title="闇を引く"
                  onClick={handleDraw}
                >
                  <Image
                    src="/hurt_dark.png"
                    alt=""
                    width={200}
                    height={200}
                    className="kirakuji-reveal -translate-y-2 h-auto w-[200px] md:w-[220px]"
                  />
                  <span className="sr-only">闇を引く</span>
                </button>
              </section>
            )}

            {currentStep === "error" && (
              <section className="flex flex-col gap-4 text-center">
                <p className="font-medium text-base text-red-600">
                  {errorMessage}
                </p>
                <button
                  className="rounded-full border border-red-200 px-6 py-3 font-semibold text-red-700 text-sm"
                  type="button"
                  onClick={() => handleRetry()}
                >
                  入力画面へ戻る
                </button>
              </section>
            )}
          </main>
        </div>
      )}
      {isTransitioning && (
        <KirakujiTransitionOverlay onAnimationComplete={completeTransition} />
      )}
    </div>
  );
}
