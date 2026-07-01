// Top-level build file. Plugins are declared here (apply false) so their
// versions come from the version catalog once, and each module opts in.
plugins {
    alias(libs.plugins.android.application) apply false
    alias(libs.plugins.kotlin.android) apply false
    alias(libs.plugins.kotlin.compose) apply false
    alias(libs.plugins.kotlin.serialization) apply false
}
