import React, { useState, useEffect } from "react";
import { Header, Footer, Navbar, LoadingSpinner } from "../components";
import { useAuth } from "../contexts/AuthContext";
import * as wfhPeriodService from "../services/wfhPeriodService";
import type { WFHPeriod, CreateWFHPeriodRequest } from "../types/wfh-period.types";
import { CreateWFHPeriodModal } from "../components/modals/CreateWFHPeriodModal";
import toast from "react-hot-toast";

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export const WFHPeriodPage: React.FC = () => {
  const { user } = useAuth();

  const [periods, setPeriods] = useState<WFHPeriod[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchPeriods = async () => {
    try {
      setIsLoading(true);
      const res = await wfhPeriodService.listWFHPeriods();
      if (res.success && res.data) {
        setPeriods(res.data);
      } else {
        setPeriods([]);
      }
    } catch (err: any) {
      const msg =
        err?.error?.message || "Failed to load WFH periods.";
      toast.error(msg);
      setPeriods([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPeriods();
  }, []);

  const handleCreate = async (payload: CreateWFHPeriodRequest) => {
    try {
      await wfhPeriodService.createWFHPeriod(payload);
      toast.success("WFH period created successfully!");
      fetchPeriods();
    } catch (err: any) {
      const msg =
        err?.error?.message || "Failed to create WFH period.";
      toast.error(msg);
      throw err;
    }
  };

  const handleDelete = async (id: string) => {
    try {
      setDeletingId(id);
      await wfhPeriodService.deleteWFHPeriod(id);
      toast.success("WFH period deleted successfully!");
      setPeriods((prev) => prev.filter((p) => p.id !== id));
    } catch (err: any) {
      const msg =
        err?.error?.message || "Failed to delete WFH period.";
      toast.error(msg);
    } finally {
      setDeletingId(null);
    }
  };

  if (isLoading) {
    return <LoadingSpinner message="Loading WFH periods…" />;
  }

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header userName={user?.name} userRole={user?.role} />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 max-w-2xl flex flex-col">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
          <div>
            <h2 className="text-4xl font-black text-left text-[var(--color-background-dark)] mb-2 tracking-tight">
              WFH Periods
            </h2>
            <p className="text-lg text-[var(--color-text-sub)] font-medium">
              Manage company-wide Work From Home periods
            </p>
          </div>

          <button
            onClick={() => setIsModalOpen(true)}
            className="px-5 py-3 rounded-xl bg-gradient-to-br from-[#fa8c47] to-[#e57a36] text-white font-bold shadow-[6px_6px_12px_#e6dccf,-6px_-6px_12px_#ffffff] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center gap-2 text-sm self-start md:self-auto"
          >
            <span className="material-symbols-outlined text-[18px]">add</span>
            Create WFH Period
          </button>
        </div>

        {periods.length > 0 ? (
          <div className="flex flex-col gap-4">
            {periods.map((period) => (
              <div
                key={period.id}
                className="bg-[#FFFDF5] rounded-[2rem] border border-[#e6dccf] p-6 shadow-[var(--shadow-clay-card)] flex flex-col md:flex-row md:items-center justify-between gap-4"
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 rounded-2xl flex items-center justify-center shadow-[var(--shadow-clay-button)] bg-white shrink-0">
                    <span className="material-symbols-outlined text-[24px] text-[var(--color-primary)]">
                      date_range
                    </span>
                  </div>
                  <div>
                    <h3 className="text-lg font-black tracking-tight text-[var(--color-background-dark)]">
                      {formatDate(period.start_date)} — {formatDate(period.end_date)}
                    </h3>
                    {period.reason && (
                      <p className="text-sm text-[var(--color-text-sub)] mt-0.5">
                        {period.reason}
                      </p>
                    )}
                    <div className="flex items-center gap-3 mt-2">
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-lg text-[10px] font-bold uppercase tracking-wider ${
                          period.active
                            ? "bg-green-50 text-green-600 border border-green-200"
                            : "bg-gray-50 text-gray-500 border border-gray-200"
                        }`}
                      >
                        {period.active ? "Active" : "Inactive"}
                      </span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => handleDelete(period.id)}
                  disabled={deletingId === period.id}
                  className="self-start md:self-center px-4 py-2 rounded-xl border border-red-200 bg-red-50 text-red-600 text-sm font-semibold hover:bg-red-100 active:shadow-[var(--shadow-clay-button-active)] transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {deletingId === period.id ? (
                    <>
                      <span className="material-symbols-outlined text-[16px] animate-spin">
                        progress_activity
                      </span>
                      Deleting…
                    </>
                  ) : (
                    <>
                      <span className="material-symbols-outlined text-[16px]">
                        delete
                      </span>
                      Delete
                    </>
                  )}
                </button>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-[#FFFDF5] border border-[#e6dccf] px-6 py-4 rounded-2xl text-sm text-[var(--color-text-sub)] flex items-center gap-3">
            <span className="material-symbols-outlined text-[20px]">info</span>
            No WFH periods found. Use the button above to create one.
          </div>
        )}
      </main>

      <Footer
        links={[
          { label: "Privacy", href: "#" },
          { label: "Terms", href: "#" },
          { label: "Support", href: "#" },
        ]}
      />

      <CreateWFHPeriodModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreate}
      />
    </div>
  );
};
