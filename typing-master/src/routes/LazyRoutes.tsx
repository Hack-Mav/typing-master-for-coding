/**
 * Lazy-loaded route components
 * Implements code splitting for better performance
 */

import { Suspense } from 'react';
import { lazyWithRetry } from '../utils/lazyLoad';

// Loading component
const LoadingFallback = () => (
  <div className="flex items-center justify-center min-h-screen">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
      <p className="text-gray-600">Loading...</p>
    </div>
  </div>
);

// Lazy load route components (only existing components)
// Components with default exports
export const ZenMode = lazyWithRetry(() => import('../components/ZenMode'));
export const TimedDrillMode = lazyWithRetry(
  () => import('../components/TimedDrillMode')
);
export const MonacoTypingInterface = lazyWithRetry(
  () => import('../components/MonacoTypingInterface')
);
export const ResultsAnalyticsInterface = lazyWithRetry(
  () => import('../components/ResultsAnalyticsInterface')
);
export const AccessibilityProvider = lazyWithRetry(
  () => import('../components/AccessibilitySettings')
);
export const AdminDashboard = lazyWithRetry(
  () => import('../components/Admin/AdminDashboard')
);
export const LessonBuilder = lazyWithRetry(
  () => import('../components/Admin/LessonBuilder')
);

// Components with named exports - need special handling
export const AssessmentMode = lazyWithRetry(() =>
  import('../components/AssessmentMode').then(module => ({
    default: module.AssessmentMode,
  }))
);
export const AccessibilitySettings = lazyWithRetry(() =>
  import('../components/AccessibilitySettings').then(module => ({
    default: module.AccessibilitySettings,
  }))
);

// Wrapper component with Suspense
export const withSuspense = (Component: React.LazyExoticComponent<any>) => {
  return (props: any) => (
    <Suspense fallback={<LoadingFallback />}>
      <Component {...props} />
    </Suspense>
  );
};

const LazyRoutes = {
  ZenMode: withSuspense(ZenMode),
  TimedDrillMode: withSuspense(TimedDrillMode),
  AssessmentMode: withSuspense(AssessmentMode),
  MonacoTypingInterface: withSuspense(MonacoTypingInterface),
  ResultsAnalyticsInterface: withSuspense(ResultsAnalyticsInterface),
  AccessibilityProvider: withSuspense(AccessibilityProvider),
  AccessibilitySettings: withSuspense(AccessibilitySettings),
  AdminDashboard: withSuspense(AdminDashboard),
  LessonBuilder: withSuspense(LessonBuilder),
};

export default LazyRoutes;
