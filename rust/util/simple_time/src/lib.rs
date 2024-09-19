//! A very simple, no_std friendly time library for Rust that provides:
//! - a Time type that wraps a i128 representing nanoseconds since the Unix epoch
//! - a Duration type that wraps a i128 representing a number of nanoseconds
//! - const fn math operations for Time and Duration, plus basic core::ops trait implementations
//!
//! Any conversion to calendar time and parsing and formatting of date/time strings
//! should be accomplished with other libraries.
//! This library just supports very simple high precision time calculations and nothing more.

#![no_std]

#[cfg(feature = "std")]
extern crate alloc;

use core::ops::{Add, Sub, Neg, AddAssign, SubAssign, Mul, MulAssign};

/// Time is as the number of nanoseconds since the Unix epoch.
/// The default value of Time is the Unix epoch 1970-01-01 00:00:00 UTC.
/// This default may not be suitable for all applications, so
/// wrap time in an Option.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Default)]
pub struct Time(i128);

impl Time {
    /// Constructs a time instance with nanoseconds since the Unix epoch.
    pub const fn from_unix_nanos(time: i128) -> Self {
        Time(time)
    }

    /// Constructs a time instance with seconds since the Unix epoch.
    pub const fn from_unix_secs(time: i64) -> Self {
        Time(time as i128 * 1_000_000_000)
    }

    /// Returns the number of nanoseconds since the Unix epoch.
    pub const fn unix_nanos(&self) -> i128 {
        self.0
    }

    /// Adds a duration to the time.
    pub const fn add(self, duration: Duration) -> Self {
        Time(self.0 + duration.0)
    }

    /// Subtracts a duration from the time.
    pub const fn sub(self, duration: Duration) -> Self {
        Time(self.0 - duration.0)
    }

    /// Returns the duration elapsed since the other time.
    pub const fn since(self, time: Time) -> Duration {
        Duration(self.0 - time.0)
    }

    /// Returns the duration until the other time.
    pub const fn until(self, time: Time) -> Duration {
        Duration(time.0 - self.0)
    }
}

/// Duration is a number of nanoseconds.
/// The default value of Duration is 0 nanoseconds.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Default)]
pub struct Duration(i128);

impl Duration {
    /// Constructs a duration instance with nanoseconds.
    pub const fn from_nanos(duration: i128) -> Self {
        Duration(duration)
    }

    /// Constructs a duration instance with seconds.
    pub const fn from_secs(duration: i64) -> Self {
        Duration(duration as i128 * 1_000_000_000)
    }

    /// Returns the number of nanoseconds in the duration.
    pub const fn nanos(&self) -> i128 {
        self.0
    }

    /// Multiplies the duration by a scalar.
    pub const fn times(self, rhs: i128) -> Self {
        Duration(self.0 * rhs)
    }

    /// One second.
    pub const SECOND: Duration = Duration(1_000_000_000);

    /// One minute.
    pub const MINUTE: Duration = Duration::SECOND.times(60);

    /// One hour.
    pub const HOUR: Duration = Duration::MINUTE.times(60);

    /// One day calculated simply as 24 hours, no DST or leap seconds.
    pub const DAY: Duration = Duration::HOUR.times(24);

    /// One week calculated simply as 7 days, no DST or leap seconds.
    pub const WEEK: Duration = Duration::DAY.times(7);
}

impl Add<Duration> for Time {
    type Output = Time;

    fn add(self, rhs: Duration) -> Self::Output {
        Time(self.0 + rhs.0)
    }
}

impl AddAssign<Duration> for Time {
    fn add_assign(&mut self, rhs: Duration) {
        self.0 += rhs.0;
    }
}

impl Sub<Duration> for Time {
    type Output = Time;

    fn sub(self, rhs: Duration) -> Self::Output {
        Time(self.0 - rhs.0)
    }
}

impl SubAssign<Duration> for Time {
    fn sub_assign(&mut self, rhs: Duration) {
        self.0 -= rhs.0;
    }
}

impl Sub<Time> for Time {
    type Output = Duration;

    fn sub(self, rhs: Time) -> Self::Output {
        Duration(self.0 - rhs.0)
    }
}

impl Neg for Duration {
    type Output = Duration;

    fn neg(self) -> Self::Output {
        Duration(-self.0)
    }
}

impl Mul<i128> for Duration {
    type Output = Duration;

    fn mul(self, rhs: i128) -> Self::Output {
        Duration(self.0 * rhs)
    }
}

impl MulAssign<i128> for Duration {
    fn mul_assign(&mut self, rhs: i128) {
        self.0 *= rhs;
    }
}