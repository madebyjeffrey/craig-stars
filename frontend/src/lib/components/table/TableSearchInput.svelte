<script lang="ts">
	export let value = '';
</script>

<input
	type="text"
	name="search"
	placeholder="search"
	class="input input-bordered input-sm w-full sm:w-auto"
	autocomplete="off"
	autocorrect="off"
	autocapitalize="off"
	spellcheck="false"
	bind:value
	{...$$restProps}
/>
