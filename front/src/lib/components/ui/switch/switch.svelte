<script lang="ts">
	import { Switch as SwitchPrimitive } from "bits-ui";
	import { cn, type WithoutChildrenOrChild } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		checked = $bindable(false),
		size = "default",
		...restProps
	}: WithoutChildrenOrChild<SwitchPrimitive.RootProps> & {
		size?: "sm" | "default";
	} = $props();
</script>

<SwitchPrimitive.Root
	bind:ref
	bind:checked
	data-slot="switch"
	data-size={size}
	class={cn(
		"shrink-0 rounded-full border-2 border-black transition-colors duration-200 outline-none peer group/switch relative inline-flex items-center after:absolute after:-inset-x-3 after:-inset-y-2 data-disabled:cursor-not-allowed data-disabled:opacity-50 data-[size=default]:h-7 data-[size=default]:w-12 data-[size=sm]:h-5 data-[size=sm]:w-9",
		"data-[state=checked]:bg-accent data-[state=unchecked]:bg-input",
		"focus-visible:ring-2 focus-visible:ring-black",
		className
	)}
	{...restProps}
>
	<SwitchPrimitive.Thumb
		data-slot="switch-thumb"
		class={cn(
			"bg-white border-2 border-black rounded-full pointer-events-none block transition-transform duration-200",
			"group-data-[size=default]/switch:size-5 group-data-[size=sm]/switch:size-3.5",
			"data-[state=checked]:translate-x-[22px] data-[state=unchecked]:translate-x-[2px]",
		)}
	/>
</SwitchPrimitive.Root>
