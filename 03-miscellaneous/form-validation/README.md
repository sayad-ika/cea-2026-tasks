# Job Application Form

A multi-step job application form built with React, TypeScript, and Tailwind CSS. Features client-side validation, conditional fields, dynamic job entries, resume upload, and a dedicated success page.

## Features

- **3-step wizard**: Personal Info → Experience → Review & Submit
- **Cream & black theme**: Warm, elegant color palette
- **Collapsible social links**: Optional LinkedIn, GitHub, and Portfolio URLs
- **Conditional experience**: Toggle "I have prior experience" to show/hide years and previous jobs
- **Dynamic job entries**: Add/remove previous jobs with "I currently work here" support
- **Resume upload**: Required file upload with `.pdf`, `.docx`, `.doc` validation
- **Success page**: Shows submitted data summary on successful submission; toast only on errors
- **Client-side validation**: Zod v4 schema with conditional refinements
- **Async email check**: Simulated uniqueness validation on blur

## Tech Stack

- React 19 + TypeScript
- Vite
- Tailwind CSS v4
- react-hook-form + @hookform/resolvers
- Zod v4

## Getting Started

```bash
# Install dependencies
pnpm install

# Start dev server
pnpm dev

# Build for production
pnpm build

# Preview production build
pnpm preview

# Lint
pnpm lint
```

## Project Structure

```
src/
  components/
    PersonalInfo.tsx      # Step 1: Name, email, phone, social links
    Experience.tsx        # Step 2: Experience toggle, resume, jobs
    ReviewSubmit.tsx      # Step 3: Review all data before submit
    SubmissionSuccess.tsx # Post-submit success page with data summary
    Toast.tsx             # Error notification toast
    FieldError.tsx        # Reusable field error message
  lib/
    schema.ts             # Zod validation schema
    mockApi.ts            # Simulated API for email check & submission
  type/
    toast.ts              # Toast type definitions
  App.tsx                 # Main app with stepper and form state
```

## Notes

- The mock API simulates a 50% failure rate on submission to demonstrate the error toast
- Three pre-existing lint warnings are unrelated to this project's changes (unused variables in scaffold code)
- File inputs reset their visual state on component remount (navigating between steps), but the selected file is preserved in form state and shown via a persistent filename label
