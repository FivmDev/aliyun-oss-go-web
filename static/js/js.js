    document.addEventListener('DOMContentLoaded', function () {
    const form = document.querySelector("form");
    const fileInput = document.querySelector("#file");
    const callbackInfo = document.querySelector("#callback-info");
    const callbackContent = document.querySelector("#callback-content");

    form.addEventListener("submit", (event) => {
    event.preventDefault();

    const file = fileInput.files[0];

    if (!file) {
    alert('请选择一个文件再上传。');
    return;
}

    const filename = file.name;

    fetch("/get_post_signature_for_oss_upload", { method: "GET" })
    .then((response) => {
    if (!response.ok) {
    throw new Error("获取签名失败");
}
    return response.json();
})
    .then((data) => {
    let formData = new FormData();
    formData.append("success_action_status", "200");
    formData.append("policy", data.policy);
    formData.append("x-oss-signature", data.signature);
    formData.append("x-oss-signature-version", "OSS4-HMAC-SHA256");
    formData.append("x-oss-credential", data.x_oss_credential);
    formData.append("x-oss-date", data.x_oss_date);
    formData.append("key", data.dir + file.name); // 文件名
    formData.append("x-oss-security-token", data.security_token);
    formData.append("callback", data.callback);  // 添加回调参数
    formData.append("file", file); // file 必须为最后一个表单域

    return fetch(data.host, {
    method: "POST",
    body: formData
});
})
    .then((response) => {
    if (response.ok) {
    console.log("上传成功");
    alert("文件已上传");
    return response.json();  // 解析回调信息
} else {
    console.log("上传失败", response);
    alert("上传失败，请稍后再试");
}
})
    .then((callbackData) => {
    if (callbackData) {
    callbackContent.textContent = JSON.stringify(callbackData, null, 2);
    callbackInfo.style.display = "block";
}
})
    .catch((error) => {
    console.error("发生错误:", error);
});
});
});