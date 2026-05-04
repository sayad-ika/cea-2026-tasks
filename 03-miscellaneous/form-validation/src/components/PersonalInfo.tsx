import { useState } from "react";
import { useFormContext } from "react-hook-form";
import type { ApplicationFormData } from "../lib/schema";
import { checkEmailUnique } from "../lib/mockApi";
import { FieldError } from "./FieldError";

export function PersonalInfo() {
    const {
        register,
        getValues,
        setValue,
        setError,
        clearErrors,
        formState: { errors, dirtyFields },
    } = useFormContext<ApplicationFormData>();

    const [checkingEmail, setCheckingEmail] = useState(false);
    const [showSocialLinks, setShowSocialLinks] = useState(
        () =>
            !!(
                getValues("linkedinUrl") ||
                getValues("githubUrl") ||
                getValues("portfolioUrl")
            ),
    );

    async function handleEmailBlur() {
        const email = getValues("email");
        if (!email || errors.email) return;

        setCheckingEmail(true);
        const isUnique = await checkEmailUnique(email);
        setCheckingEmail(false);

        if (!isUnique) {
            setError("email", {
                type: "manual",
                message: "This email is already registered",
            });
        } else {
            clearErrors("email");
        }
    }

    function handleToggleSocialLinks() {
        const willCollapse = showSocialLinks;
        setShowSocialLinks(!showSocialLinks);
        if (willCollapse) {
            setValue("linkedinUrl", "");
            setValue("githubUrl", "");
            setValue("portfolioUrl", "");
            clearErrors("linkedinUrl");
            clearErrors("githubUrl");
            clearErrors("portfolioUrl");
        }
    }

    return (
        <fieldset className="space-y-5">
            <legend className="text-xl font-semibold text-neutral-900 mb-4">
                Personal Information
            </legend>

            <div>
                <label
                    htmlFor="firstName"
                    className="block text-sm font-medium text-neutral-700"
                >
                    First Name{" "}
                    <span aria-hidden="true" className="text-red-500">
                        *
                    </span>
                </label>
                <input
                    id="firstName"
                    {...register("firstName")}
                    aria-invalid={!!errors.firstName}
                    aria-describedby="firstName-error"
                    className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
            shadow-sm focus:outline-none focus:ring-2
            ${
                errors.firstName
                    ? "border-red-500 focus:ring-red-300"
                    : "border-[#D6CDBF] focus:ring-neutral-400"
            }`}
                />
                <FieldError
                    id="firstName-error"
                    message={errors.firstName?.message}
                />
            </div>

            <div>
                <label
                    htmlFor="lastName"
                    className="block text-sm font-medium text-neutral-700"
                >
                    Last Name{" "}
                    <span aria-hidden="true" className="text-red-500">
                        *
                    </span>
                </label>
                <input
                    id="lastName"
                    {...register("lastName")}
                    aria-invalid={!!errors.lastName}
                    aria-describedby="lastName-error"
                    className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
            shadow-sm focus:outline-none focus:ring-2
            ${
                errors.lastName
                    ? "border-red-500 focus:ring-red-300"
                    : "border-[#D6CDBF] focus:ring-neutral-400"
            }`}
                />
                <FieldError
                    id="lastName-error"
                    message={errors.lastName?.message}
                />
            </div>

            <div>
                <label
                    htmlFor="email"
                    className="block text-sm font-medium text-neutral-700"
                >
                    Email{" "}
                    <span aria-hidden="true" className="text-red-500">
                        *
                    </span>
                </label>
                <div className="relative">
                    <input
                        id="email"
                        type="email"
                        {...register("email", {
                            onChange: dirtyFields.email ? undefined : undefined,
                        })}
                        onBlur={handleEmailBlur}
                        aria-invalid={!!errors.email}
                        aria-describedby="email-error email-hint"
                        className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
              shadow-sm focus:outline-none focus:ring-2 pr-10
              ${
                  errors.email
                      ? "border-red-500 focus:ring-red-300"
                      : "border-[#D6CDBF] focus:ring-neutral-400"
              }`}
                    />
                    {checkingEmail && (
                        <span
                            aria-hidden="true"
                            className="absolute right-3 top-1/2 -translate-y-1/2
                         h-4 w-4 rounded-full border-2 border-neutral-900
                         border-t-transparent animate-spin"
                        />
                    )}
                </div>
                <p id="email-hint" className="mt-1 text-xs text-neutral-500">
                    We'll use this to contact you about your application.
                </p>
                <FieldError id="email-error" message={errors.email?.message} />
            </div>

            <div>
                <label
                    htmlFor="phone"
                    className="block text-sm font-medium text-neutral-700"
                >
                    Phone{" "}
                    <span aria-hidden="true" className="text-red-500">
                        *
                    </span>
                </label>
                <input
                    id="phone"
                    type="tel"
                    placeholder="+1 (555) 123-4567"
                    {...register("phone")}
                    aria-invalid={!!errors.phone}
                    aria-describedby="phone-error"
                    className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
            shadow-sm focus:outline-none focus:ring-2
            ${
                errors.phone
                    ? "border-red-500 focus:ring-red-300"
                    : "border-[#D6CDBF] focus:ring-neutral-400"
            }`}
                />
                <FieldError id="phone-error" message={errors.phone?.message} />
            </div>

            <div>
                <button
                    type="button"
                    onClick={handleToggleSocialLinks}
                    className="text-sm font-medium text-neutral-900 hover:text-neutral-700 underline"
                >
                    {showSocialLinks
                        ? "- Hide social links"
                        : "+ Add social links (optional)"}
                </button>
            </div>

            {showSocialLinks && (
                <div className="space-y-5">
                    <div>
                        <label
                            htmlFor="linkedinUrl"
                            className="block text-sm font-medium text-neutral-700"
                        >
                            LinkedIn{" "}
                            <span className="text-neutral-400 font-normal">
                                (optional)
                            </span>
                        </label>
                        <input
                            id="linkedinUrl"
                            type="url"
                            placeholder="https://linkedin.com/in/your-profile"
                            {...register("linkedinUrl")}
                            aria-invalid={!!errors.linkedinUrl}
                            aria-describedby="linkedinUrl-error"
                            className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                shadow-sm focus:outline-none focus:ring-2
                ${
                    errors.linkedinUrl
                        ? "border-red-500 focus:ring-red-300"
                        : "border-[#D6CDBF] focus:ring-neutral-400"
                }`}
                        />
                        <FieldError
                            id="linkedinUrl-error"
                            message={errors.linkedinUrl?.message}
                        />
                    </div>

                    <div>
                        <label
                            htmlFor="githubUrl"
                            className="block text-sm font-medium text-neutral-700"
                        >
                            GitHub{" "}
                            <span className="text-neutral-400 font-normal">
                                (optional)
                            </span>
                        </label>
                        <input
                            id="githubUrl"
                            type="url"
                            placeholder="https://github.com/your-username"
                            {...register("githubUrl")}
                            aria-invalid={!!errors.githubUrl}
                            aria-describedby="githubUrl-error"
                            className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                shadow-sm focus:outline-none focus:ring-2
                ${
                    errors.githubUrl
                        ? "border-red-500 focus:ring-red-300"
                        : "border-[#D6CDBF] focus:ring-neutral-400"
                }`}
                        />
                        <FieldError
                            id="githubUrl-error"
                            message={errors.githubUrl?.message}
                        />
                    </div>

                    <div>
                        <label
                            htmlFor="portfolioUrl"
                            className="block text-sm font-medium text-neutral-700"
                        >
                            Portfolio{" "}
                            <span className="text-neutral-400 font-normal">
                                (optional)
                            </span>
                        </label>
                        <input
                            id="portfolioUrl"
                            type="url"
                            placeholder="https://your-portfolio.com"
                            {...register("portfolioUrl")}
                            aria-invalid={!!errors.portfolioUrl}
                            aria-describedby="portfolioUrl-error"
                            className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                shadow-sm focus:outline-none focus:ring-2
                ${
                    errors.portfolioUrl
                        ? "border-red-500 focus:ring-red-300"
                        : "border-[#D6CDBF] focus:ring-neutral-400"
                }`}
                        />
                        <FieldError
                            id="portfolioUrl-error"
                            message={errors.portfolioUrl?.message}
                        />
                    </div>
                </div>
            )}
        </fieldset>
    );
}
