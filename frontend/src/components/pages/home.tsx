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
  const contentLength = content.length;
  const trimmedLength = content.trim().length;
  const isSubmitDisabled = trimmedLength === 0 || contentLength > 140;
  const isSubmitButtonDisabled = isSubmitDisabled || currentStep === "loading";

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
    <div className="relative flex min-h-screen items-end justify-center overflow-hidden bg-zinc-50 px-4 py-8 font-sans text-zinc-900 md:px-0">
      <div className="pointer-events-none absolute inset-x-0 top-0 z-10 flex justify-center">
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
      <div className="relative z-20 flex w-full max-w-xl flex-col items-center gap-4">
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
          <Image
            src="/hurt_normal.png"
            alt=""
            width={220}
            height={220}
            className="h-auto w-[200px] md:w-[220px]"
            priority
          />
          <span className="sr-only">闇を投げる</span>
        </button>
      </div>

      {isModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4 py-8"
          role="dialog"
          aria-modal="true"
          onPointerDown={() => {
            setIsModalOpen(false);
            handleRetry();
          }}
        >
          <main
            className="relative flex w-full max-w-lg flex-col gap-8 rounded-3xl bg-white px-6 py-10 shadow-lg md:max-w-xl md:px-8 md:py-12"
            ref={modalRef}
            onPointerDown={(event) => event.stopPropagation()}
          >
            <button
              className="absolute top-4 right-4 rounded-full bg-zinc-100 px-3 py-1 font-semibold text-xs text-zinc-600 hover:bg-zinc-200"
              type="button"
              onClick={() => {
                setIsModalOpen(false);
                handleRetry();
              }}
            >
              閉じる
            </button>

            {(currentStep === "input" ||
              (currentStep === "loading" && loadingOrigin === "input")) && (
              <section className="flex flex-col gap-4">
                <textarea
                  className="min-h-[140px] w-full resize-none rounded-2xl border border-zinc-200 px-4 py-3 text-sm outline-none disabled:cursor-not-allowed disabled:bg-zinc-100"
                  maxLength={140}
                  placeholder="ここに闇を投げる"
                  value={content}
                  onChange={handleContentChange}
                  ref={inputRef}
                  disabled={currentStep === "loading"}
                />
                <button
                  className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white disabled:cursor-not-allowed disabled:bg-zinc-400"
                  type="button"
                  onClick={handleSubmit}
                  disabled={isSubmitButtonDisabled}
                >
                  懺悔する
                </button>
              </section>
            )}

            {currentStep === "ready" && (
              <section className="flex flex-col gap-4 text-center">
                <p className="font-medium text-base">
                  お告げを捧げました。きらくじを受け取りますか？
                </p>
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
                    className="h-auto w-[180px] md:w-[200px]"
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
