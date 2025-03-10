<template>
    {{ url }}
    </template>
    
    <script setup lang="ts">
    import { qqLoginURL } from '@/api/base';
    import { ref } from 'vue';
    
    const url = ref('')
    
    const getUrl = async ()=>{
        const res = await qqLoginURL()
        if (res.code===0){
            url.value=res.data
        }
    }
    
    getUrl()
    </script>
    
    <style scoped lang="scss">
    
    </style>
    