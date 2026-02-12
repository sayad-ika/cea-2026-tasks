import React, { useEffect, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import {
  Header,
  Footer,
  EmployeeMenuCard,
  LoadingSpinner,
  Navbar,
  WorkLocationDateModal,
} from "../components";
import type { MealType as MealTypeEnum } from "../types";
import type { MealType } from "../components/cards/EmployeeMenuCard";
import * as mealService from "../services/mealService";
import * as userService from "../services/userService";
import * as workLocationService from "../services/workLocationService";
import type { WorkLocation } from "../services/workLocationService";
import toast from "react-hot-toast";

const toTitleCase = (value: string) =>
  value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

const normalizeWorkLocation = (value: string | null | undefined) => {
  if (!value) return null;
  const normalized = value.trim().toLowerCase();
  if (normalized === "office" || normalized === "wfh") {
    return normalized as WorkLocation;
  }
  return null;
};

const getErrorMessage = (error: unknown, fallback: string) => {
  if (typeof error === "object" && error !== null) {
    const candidate = error as {
      error?: { message?: string };
      message?: string;
    };

    if (candidate.error?.message) {
      return candidate.error.message;
    }

    if (candidate.message) {
      return candidate.message;
    }
  }

  return fallback;
};

export const Home: React.FC = () => {
  const { user } = useAuth();
  const [meals, setMeals] = useState<MealType[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [dashboardDate, setDashboardDate] = useState("");
  const [workLocationDate, setWorkLocationDate] = useState("");
  const [workLocation, setWorkLocation] = useState<WorkLocation | null>(null);
  const [teamName, setTeamName] = useState<string | null>(null);
  const [isUpdatingLocation, setIsUpdatingLocation] = useState(false);
  const [showLocationDateModal, setShowLocationDateModal] = useState(false);

  useEffect(() => {
    const fetchDashboardData = async () => {
      setIsLoading(true);
      setError("");
      const todayDate = new Date().toISOString().slice(0, 10);

      const [mealsResult, locationResult] = await Promise.allSettled([
        mealService.getTodaysMeals(),
        workLocationService.getMyWorkLocation(todayDate),
      ]);

      let resolvedDate = "";

      // Fetch team assignment strictly for display
      try {
        const teamRes = await userService.getTeamAssignment();
        if (teamRes.success && teamRes.data && teamRes.data.length > 0) {
          setTeamName(teamRes.data[0].team_name);
        }
      } catch (err) {
        console.error("Failed to fetch team assignment", err);
        // Do not block page load for this
      }

      if (mealsResult.status === "fulfilled") {
        const response = mealsResult.value;
        if (response.success && response.data) {
          resolvedDate = response.data.date;
          const mealsWithStatus: MealType[] = response.data.available_meals.map(
            (mealType) => {
              const participation = response.data.participations.find(
                (entry) => entry.meal_type === mealType,
              );

              return {
                meal_type: mealType,
                is_participating: participation?.is_participating ?? true,
                opted_out_at: null,
              };
            },
          );

          setMeals(mealsWithStatus);
        } else {
          setMeals([]);
          setError("Meals are currently unavailable.");
        }
      } else {
        const msg = getErrorMessage(
          mealsResult.reason,
          "Failed to load meals.",
        );
        setMeals([]);
        setError(msg);
        toast.error(msg);
      }

      if (locationResult.status === "fulfilled") {
        setWorkLocation(normalizeWorkLocation(locationResult.value.location));
        // setWorkLocationSource(locationResult.value.source);
        setWorkLocationDate(locationResult.value.date);
        if (!resolvedDate) {
          resolvedDate = locationResult.value.date;
        }
      } else {
        const msg = getErrorMessage(
          locationResult.reason,
          "Failed to load work location status.",
        );
        toast.error(msg);
      }

      if (resolvedDate) {
        setDashboardDate(resolvedDate);
      }

      setIsLoading(false);
    };

    fetchDashboardData();
  }, []);

  const handleMealToggle = async (mealType: string) => {
    const mealTypeEnum = mealType as MealTypeEnum;
    const currentMeal = meals.find((meal) => meal.meal_type === mealType);
    if (!currentMeal) return;

    const newStatus = !currentMeal.is_participating;

    setMeals((prev) =>
      prev.map((meal) =>
        meal.meal_type === mealType
          ? { ...meal, is_participating: newStatus }
          : meal,
      ),
    );

    try {
      await mealService.setMealParticipation({
        date: dashboardDate,
        meal_type: mealTypeEnum,
        participating: newStatus,
      });
      toast.success(
        `${toTitleCase(mealType)} ${newStatus ? "opted in" : "opted out"}.`,
      );
    } catch (err: unknown) {
      setMeals((prev) =>
        prev.map((meal) =>
          meal.meal_type === mealType
            ? { ...meal, is_participating: !newStatus }
            : meal,
        ),
      );

      const msg = getErrorMessage(err, "Failed to update participation.");
      setError(msg);
      toast.error(msg);
    }
  };

  const handleWorkLocationChange = async (nextLocation: WorkLocation) => {
    if (isUpdatingLocation || workLocation === nextLocation) return;

    const targetDate =
      workLocationDate ||
      dashboardDate ||
      new Date().toISOString().slice(0, 10);
    const previousLocation = workLocation;

    setWorkLocation(nextLocation);
    setIsUpdatingLocation(true);

    try {
      await workLocationService.updateMyWorkLocation({
        date: targetDate,
        location: nextLocation,
      });
      setWorkLocationDate(targetDate);
      // setWorkLocationSource("explicit");
      toast.success(`Work location set to ${nextLocation.toUpperCase()}.`);
    } catch (err: unknown) {
      setWorkLocation(previousLocation);
      const msg = getErrorMessage(err, "Failed to update work location.");
      setError(msg);
      toast.error(msg);
    } finally {
      setIsUpdatingLocation(false);
    }
  };

  const handleDateLocationUpdate = async (
    date: string,
    location: WorkLocation
  ) => {
    try {
      await workLocationService.updateMyWorkLocation({
        date,
        location,
      });
      toast.success(
        `Work location set to ${location.toUpperCase()} for ${new Date(date).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}.`
      );

      // If the selected date is today, update the current work location state
      const todayDate = new Date().toISOString().slice(0, 10);
      if (date === todayDate) {
        setWorkLocation(location);
        setWorkLocationDate(date);
      }
    } catch (err: unknown) {
      const msg = getErrorMessage(err, "Failed to update work location.");
      setError(msg);
      toast.error(msg);
      throw err; // Re-throw to let modal handle the error state
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      weekday: "long",
      month: "short",
      day: "numeric",
    });
  };

  const displayDate = dashboardDate || workLocationDate;

  if (isLoading) {
    return <LoadingSpinner message="Loading your dashboard..." />;
  }

  return (
    <div className="font-display bg-[var(--color-background-light)] text-[var(--color-text-main)] min-h-screen flex flex-col overflow-x-hidden">
      <Header
        userName={user?.name || "User"}
        userRole={user?.role || "employee"}
      />
      <Navbar />

      <main className="flex-grow container mx-auto px-6 py-8 md:px-12 flex flex-col">
        <div className="mb-8 text-center md:text-left">
          <h2 className="text-4xl md:text-5xl font-black text-[var(--color-background-dark)] mb-2 tracking-tight">
            Today&apos;s Update
          </h2>
          <p className="text-lg text-[var(--color-text-sub)] font-medium">
            Manage your daily meals and work location for{" "}
            <span className="text-[var(--color-primary)] font-bold">
              {displayDate ? formatDate(displayDate) : "Today"}
            </span>
          </p>
        </div>

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-6 py-4 rounded-2xl text-sm">
            {error}
          </div>
        )}

        {/* Team Assignment Badge */}
        <div className="mb-10 text-center md:text-left">
          {teamName && (
            <div className="inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-orange-50 to-orange-100 text-orange-700 rounded-2xl border border-orange-200 shadow-sm mb-4">
              <span className="material-symbols-outlined text-[20px] leading-none">
                groups
              </span>
              <span className="text-sm font-bold tracking-wide">
                {teamName}
              </span>
            </div>
          )}
        </div>

        <div className="mb-12 bg-[var(--color-background-light)] rounded-3xl p-6 md:p-8 flex flex-col md:flex-row items-center justify-between gap-6 border-2 border-white/50 shadow-[var(--shadow-clay)]">
          <div className="w-full md:w-auto">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-[var(--color-primary)]/10 flex items-center justify-center text-[var(--color-primary)] shadow-[var(--shadow-clay-inset)]">
                <span className="material-symbols-outlined text-2xl">
                  location_on
                </span>
              </div>
              <div>
                <h3 className="text-xl font-bold text-left text-[var(--color-background-dark)]">
                  Work Location
                </h3>
                <p className="text-sm text-[var(--color-text-sub)]">
                  Where will you be working from today?
                </p>
              </div>
            </div>
            {/* {workLocationSource && (
              <p className="inline-flex mt-3 ml-1 text-xs font-semibold text-[var(--color-text-sub)] bg-white/60 px-3 py-1 rounded-xl border border-white/70">
                Source: {toTitleCase(workLocationSource)}
              </p>
            )} */}
          </div>

          <div className="flex flex-col md:flex-row items-center gap-4 md:gap-6 w-full md:w-auto">
            <div className="flex items-center gap-4 w-full md:w-auto">
              <button
                type="button"
                onClick={() => handleWorkLocationChange("office")}
                disabled={isUpdatingLocation}
                className={`flex-1 md:flex-none flex items-center justify-center gap-3 px-8 py-4 rounded-2xl min-w-[160px] font-bold text-sm uppercase tracking-wider relative overflow-hidden transition-all duration-300 group disabled:opacity-70 disabled:cursor-not-allowed ${
                  workLocation === "office"
                    ? "bg-gradient-to-r from-[var(--color-primary)] to-[#fb923c] text-white shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                    : "bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                }`}
              >
                <span
                  className={`absolute inset-0 bg-white/20 transition-transform duration-300 ${
                    workLocation === "office"
                      ? "translate-y-full group-hover:translate-y-0"
                      : "opacity-0"
                  }`}
                />
                <span className="relative z-10 text-2xl drop-shadow-sm">🏢</span>
                <span className="relative z-10">Office</span>
                {workLocation === "office" && (
                  <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-white animate-pulse" />
                )}
              </button>

              <button
                type="button"
                onClick={() => handleWorkLocationChange("wfh")}
                disabled={isUpdatingLocation}
                className={`flex-1 md:flex-none flex items-center justify-center gap-3 px-8 py-4 rounded-2xl min-w-[160px] font-bold text-sm uppercase tracking-wider relative overflow-hidden transition-all duration-300 group disabled:opacity-70 disabled:cursor-not-allowed ${
                  workLocation === "wfh"
                    ? "bg-gradient-to-r from-[var(--color-primary)] to-[#fb923c] text-white shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                    : "bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
                }`}
              >
                <span
                  className={`absolute inset-0 bg-white/20 transition-transform duration-300 ${
                    workLocation === "wfh"
                      ? "translate-y-full group-hover:translate-y-0"
                      : "opacity-0"
                  }`}
                />
                <span className="relative z-10 text-2xl drop-shadow-sm">🏡</span>
                <span className="relative z-10">WFH</span>
                {workLocation === "wfh" && (
                  <span className="absolute top-2 right-2 w-2 h-2 rounded-full bg-white animate-pulse" />
                )}
              </button>
            </div>

            <button
              type="button"
              onClick={() => setShowLocationDateModal(true)}
              disabled={isUpdatingLocation}
              className="w-full md:w-auto flex items-center justify-center gap-2 px-6 py-4 rounded-2xl bg-[var(--color-background-light)] text-[var(--color-text-main)] hover:text-[var(--color-primary)] font-bold text-sm uppercase tracking-wider transition-all duration-300 group disabled:opacity-70 disabled:cursor-not-allowed shadow-[var(--shadow-clay-button)] active:shadow-[var(--shadow-clay-button-active)]"
            >
              <span className="material-symbols-outlined text-xl group-hover:scale-110 transition-transform">
                calendar_month
              </span>
              <span className="relative z-10">Schedule</span>
            </button>
          </div>
        </div>

        {isUpdatingLocation && (
          <div className="mb-8 text-sm font-medium text-[var(--color-text-sub)] inline-flex items-center gap-2 self-center md:self-start">
            <span className="material-symbols-outlined animate-spin text-base">
              progress_activity
            </span>
            Updating work location...
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
          {meals.length > 0 ? (
            meals.map((meal) => (
              <EmployeeMenuCard
                key={meal.meal_type}
                meal={meal}
                onToggle={handleMealToggle}
              />
            ))
          ) : (
            <div className="col-span-full text-center py-12">
              <span className="material-symbols-outlined text-6xl text-[var(--color-text-sub)] mb-4 block">
                no_meals
              </span>
              <p className="text-[var(--color-text-sub)] text-lg">
                No meals available for today.
              </p>
            </div>
          )}
        </div>

        {/* <div className="flex flex-col md:flex-row justify-center items-center gap-6 pb-8">
          <button
            type="button"
            onClick={() => toast("Calendar view is coming soon.")}
            className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
            style={{ boxShadow: "var(--shadow-clay-button)" }}
          >
            <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
              calendar_month
            </span>
            <span className="font-bold text-sm uppercase tracking-wider">
              View Calendar
            </span>
          </button>

          <button
            type="button"
            onClick={() => toast("History view is coming soon.")}
            className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
            style={{ boxShadow: "var(--shadow-clay-button)" }}
          >
            <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
              history
            </span>
            <span className="font-bold text-sm uppercase tracking-wider">
              History
            </span>
          </button>

          <button
            type="button"
            onClick={() => toast("Preferences are coming soon.")}
            className="flex items-center gap-3 px-8 py-4 bg-[var(--color-background-light)] rounded-2xl text-[var(--color-text-main)] hover:text-[var(--color-primary)] transition-colors min-w-[200px] justify-center group"
            style={{ boxShadow: "var(--shadow-clay-button)" }}
          >
            <span className="material-symbols-outlined group-hover:scale-110 transition-transform">
              settings
            </span>
            <span className="font-bold text-sm uppercase tracking-wider">
              Preferences
            </span>
          </button>
        </div> */}
      </main>

      <Footer
        links={[
          { label: "Privacy", href: "#" },
          { label: "Terms", href: "#" },
          { label: "Support", href: "#" },
        ]}
      />

      {/* Work Location Date Modal */}
      <WorkLocationDateModal
        isOpen={showLocationDateModal}
        onClose={() => setShowLocationDateModal(false)}
        onConfirm={handleDateLocationUpdate}
        currentLocation={workLocation}
      />
    </div>
  );
};
