# kotlinx.serialization keeps generated serializers; the plugin emits the
# necessary keep rules, but guard the @Serializable data models explicitly in
# case shrinking is enabled later.
-keepclassmembers class com.phasionary.app.data.model.** {
    *** Companion;
}
-keepclasseswithmembers class com.phasionary.app.data.model.** {
    kotlinx.serialization.KSerializer serializer(...);
}
