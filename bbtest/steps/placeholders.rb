
placeholder :activity do
  match(/(active|pasive)/) do |activity|
    activity == "active"
  end
end

placeholder :side do
  match(/(credit|debit)/) do |side|
    side
  end
end

placeholder :amount do
  match(/-?\d{1,100}\.\d{1,100}|-?\d{1,100}/) do |amount|
    amount
  end
end

placeholder :count do
  match(/\d{1,100}/) do |count|
    count.to_i
  end
end

placeholder :path do
  match(/((?:\/[a-z0-9]+[a-z0-9(\/)(\-)]{1,100}[\w,\s-]+(\.?[A-Za-z0-9_-]{0,100})+))/) do |path|
    path
  end
end

placeholder :permissions do
  match(/-?[r-][w-][x-][r-][w-][x-][r-][w-][x-]/) do |permissions|
    perm = permissions.reverse[0...9].reverse.chars.each_slice(3).map { |part|
      (part[0] == 'r' ? 4 : 0) +
      (part[1] == 'w' ? 2 : 0) +
      (part[2] == 'x' ? 1 : 0)
    }.join('')

    "0#{perm}"
  end
end

placeholder :http_method do
  match(/(GET|get|POST|post|PATCH|patch|DELETE|delete)/) do |http_method|
    http_method.downcase
  end
end

placeholder :transaction_status do
  match(/(committed|rollbacked|promised)/) do |transaction_status|
    transaction_status
  end
end

placeholder :http_status do
  match(/\d{3}/) do |http_status|
    http_status.to_i
  end
end

placeholder :url do
  match(/https?:\/\/[\S]+/) do |url|
    url
  end
end

placeholder :account do
  match(/[\S]+\/[\S]+/) do |account|
    account
  end
end
