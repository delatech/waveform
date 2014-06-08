require 'json'
THRESHOLD=0.1

original_data = JSON.parse(File.read('origin.json'))
result = JSON.parse(File.read('test.json'))['Waves']
count = 0
original_data.each_with_index do |value, index|
  diff = value - result[index]
  if diff.abs > THRESHOLD
    puts "[%d] %0.5f | %0.5f" % [index, value, result[index]]
    count += 1
  end
end

puts "total values : %d" % original_data.count
puts "wrong values : %d" % count
